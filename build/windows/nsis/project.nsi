Unicode true
ManifestDPIAware true

!include "wails_tools.nsh"
!include "MUI.nsh"
!include "nsDialogs.nsh"

!ifndef ARG_WAILS_CONFIGS_DIR
    !define ARG_WAILS_CONFIGS_DIR "..\..\..\configs"
!endif

Name "${INFO_PRODUCTNAME}"
OutFile "..\..\..\bin\${INFO_PROJECTNAME}-${ARCH}-installer.exe"
!if "${WAILS_INSTALL_SCOPE}" == "user"
    InstallDir "$LOCALAPPDATA\Programs\${INFO_PRODUCTNAME}"
!else
    InstallDir "$PROGRAMFILES64\${INFO_COMPANYNAME}\${INFO_PRODUCTNAME}"
!endif

# Version information (4 parts required).
VIProductVersion "${INFO_PRODUCTVERSION}.0"
VIFileVersion    "${INFO_PRODUCTVERSION}.0"
VIAddVersionKey "CompanyName"     "${INFO_COMPANYNAME}"
VIAddVersionKey "FileDescription" "${INFO_PRODUCTNAME} Installer"
VIAddVersionKey "ProductVersion"  "${INFO_PRODUCTVERSION}"
VIAddVersionKey "FileVersion"     "${INFO_PRODUCTVERSION}"
VIAddVersionKey "LegalCopyright"  "${INFO_COPYRIGHT}"
VIAddVersionKey "ProductName"     "${INFO_PRODUCTNAME}"

# Interface.
ShowInstDetails show

!define MUI_ICON "..\icon.ico"
!define MUI_UNICON "..\icon.ico"
!define MUI_ABORTWARNING
!define MUI_FINISHPAGE_NOAUTOCLOSE
!define MUI_FINISHPAGE_RUN "$INSTDIR\${PRODUCT_EXECUTABLE}"
!define MUI_FINISHPAGE_RUN_TEXT "Launch ${INFO_PRODUCTNAME}"

!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH

Var DeleteLocalUserDataCheckbox
Var DeleteLocalUserData
Var SavedLocalAppData
Var SavedAppData

UninstPage custom un.LocalUserDataPage un.LocalUserDataPageLeave
!insertmacro MUI_UNPAGE_INSTFILES

!insertmacro MUI_LANGUAGE "English"

# Signing hooks.
#!uninstfinalize 'signtool --file "%1"'
#!finalize 'signtool --file "%1"'

# Installer callbacks.
Function .onInit
    !insertmacro wails.checkArchitecture

    SetRegView 64
    !if "${WAILS_INSTALL_SCOPE}" == "user"
        ReadRegStr $0 HKCU "${UNINST_KEY}" "InstallLocation"
    !else
        ReadRegStr $0 HKLM "${UNINST_KEY}" "InstallLocation"
    !endif
    ${If} $0 != ""
        StrCpy $INSTDIR $0
    ${EndIf}
FunctionEnd

# Uninstaller callbacks.
Function un.onInit
    ${IfNot} ${FileExists} "$INSTDIR\${PRODUCT_EXECUTABLE}"
        Return
    ${EndIf}

    ClearErrors
    FileOpen $0 "$INSTDIR\${PRODUCT_EXECUTABLE}" a
    ${IfNot} ${Errors}
        FileClose $0
        Return
    ${EndIf}

    SetErrorLevel 2
    MessageBox MB_OK|MB_ICONEXCLAMATION "${INFO_PRODUCTNAME} is running, please close it first."
    Abort
FunctionEnd

Function un.LocalUserDataPage
    nsDialogs::Create 1018
    Pop $0
    ${If} $0 == error
        Abort
    ${EndIf}

    ${NSD_CreateLabel} 0 0 100% 64u "Please choose whether to delete local user data.$\r$\n\
    If checked, all local user data will be permanently deleted and cannot be recovered.$\r$\n\
    If unchecked, all local user data will be retained on this computer and may be used again after reinstallation."
    Pop $1

    ${NSD_CreateCheckbox} 0 72u 100% 12u "Delete local user data"
    Pop $DeleteLocalUserDataCheckbox

    nsDialogs::Show
FunctionEnd

Function un.LocalUserDataPageLeave
    ${NSD_GetState} $DeleteLocalUserDataCheckbox $DeleteLocalUserData
FunctionEnd

# Installer section.
Section "Install"
    ${If} ${FileExists} "$INSTDIR\${PRODUCT_EXECUTABLE}"
        ClearErrors
        FileOpen $0 "$INSTDIR\${PRODUCT_EXECUTABLE}" a
        ${If} ${Errors}
            MessageBox MB_OK|MB_ICONEXCLAMATION "${INFO_PRODUCTNAME} is running, please close it first."
            SetErrorLevel 2
            Abort
        ${EndIf}
        FileClose $0
    ${EndIf}

    !insertmacro wails.setShellContext
    !insertmacro wails.webview2runtime

    SetOutPath $INSTDIR
    !insertmacro wails.files

    SetOverwrite off
    SetOutPath "$INSTDIR\configs"
    File /r "${ARG_WAILS_CONFIGS_DIR}\*.*"
    SetOverwrite on
    SetOutPath $INSTDIR

    CreateShortcut "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}"
    CreateShortcut "$DESKTOP\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}"

    !insertmacro wails.associateFiles
    !insertmacro wails.associateCustomProtocols
    !insertmacro wails.writeUninstaller

    WriteRegStr   SHELL_CONTEXT "${UNINST_KEY}" "InstallLocation" "$INSTDIR"
    WriteRegStr   SHELL_CONTEXT "${UNINST_KEY}" "URLInfoAbout" "https://github.com/Wexyuan/kairos"
    WriteRegDWORD SHELL_CONTEXT "${UNINST_KEY}" "NoModify" 1
    WriteRegDWORD SHELL_CONTEXT "${UNINST_KEY}" "NoRepair" 1
SectionEnd

# Uninstaller section.
Section "Uninstall"
    StrCpy $SavedLocalAppData "$LOCALAPPDATA"
    StrCpy $SavedAppData "$APPDATA"
    !insertmacro wails.setShellContext

    ${If} $DeleteLocalUserData == ${BST_CHECKED}
        RMDir /r "$SavedLocalAppData\${INFO_PROJECTNAME}"
    ${EndIf}

    RMDir /r "$SavedAppData\${PRODUCT_EXECUTABLE}"

    RMDir /r $INSTDIR
    ${If} ${FileExists} "$INSTDIR\${PRODUCT_EXECUTABLE}"
        MessageBox MB_ICONEXCLAMATION "Could not delete ${PRODUCT_EXECUTABLE}. Close the app and try again."
    ${EndIf}

    !if "${WAILS_INSTALL_SCOPE}" == "machine"
        RMDir "$PROGRAMFILES64\${INFO_COMPANYNAME}"
    !endif

    Delete "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk"
    Delete "$DESKTOP\${INFO_PRODUCTNAME}.lnk"

    !insertmacro wails.unassociateFiles
    !insertmacro wails.unassociateCustomProtocols
    !insertmacro wails.deleteUninstaller
SectionEnd
