[CmdletBinding()]
param(
    # Point to the repository root containing build/, bin/ and configs/.
    [Parameter(Mandatory = $true)][string]$RootDir,
    # Name the project passed to the build asset generator.
    [Parameter(Mandatory = $true)][string]$AppName,
    # Set the target architecture used in artifact names.
    [Parameter(Mandatory = $true)][string]$Arch,
    # Pin the release version (x.y.z) so builds never reuse a stale value.
    [Parameter(Mandatory = $true)][string]$Version,
    # Pick the pipeline stage: sync rewrites build assets, collect gathers artifacts.
    [Parameter(Mandatory = $true)][ValidateSet('sync', 'collect')][string]$Phase
)

$ErrorActionPreference = 'Stop'

# Write text back while keeping the original BOM state.
function Set-TextPreservingEncoding {
    param([string]$Path, [string]$Content)

    $bytes = [IO.File]::ReadAllBytes($Path)
    $hasBom = $bytes.Length -ge 3 -and $bytes[0] -eq 0xEF -and $bytes[1] -eq 0xBB -and $bytes[2] -eq 0xBF
    [IO.File]::WriteAllText($Path, $Content, (New-Object System.Text.UTF8Encoding($hasBom)))
}

# Replace the first regex group value while leaving surrounding text intact.
function Set-FirstGroupValue {
    param([string]$Path, [string]$Pattern, [string]$Value)

    $content = [IO.File]::ReadAllText($Path)
    $match = [regex]::new($Pattern).Match($content)
    if (-not $match.Success) {
        throw "Cannot find the target field in $Path"
    }

    $group = $match.Groups[1]
    $updated = $content.Substring(0, $group.Index) + $Value + $content.Substring($group.Index + $group.Length)
    Set-TextPreservingEncoding -Path $Path -Content $updated
}

# Stop early when a generated asset did not pick up the requested version.
function Assert-VersionPattern {
    param([string]$Path, [string]$Pattern, [string]$Expected, [switch]$Optional)

    if (-not (Test-Path $Path)) {
        if ($Optional) { return }
        throw "Missing build asset: $Path"
    }
    if (-not ([IO.File]::ReadAllText($Path) -match $Pattern)) {
        throw "Version $Expected is missing in $Path"
    }
}

$Version = $Version.Trim()
if ($Version -notmatch '^\d+\.\d+\.\d+$') {
    throw "Invalid VERSION '$Version', expected the x.y.z format"
}

$binDir = Join-Path $RootDir 'bin'
$releaseDir = Join-Path $binDir "release/v$Version"

if ($Phase -eq 'sync') {
    $configPath = Join-Path $RootDir 'build/config.yml'
    $content = [IO.File]::ReadAllText($configPath)
    $current = [regex]::Match($content, '(?m)^\s*version:\s*"([^"]+)"').Groups[1].Value
    if (-not $current) {
        throw "Cannot read the product version from $configPath"
    }

    if ($current -ne $Version) {
        # Update the source config first so the generator below picks up the new version.
        Set-FirstGroupValue -Path $configPath -Pattern '(?m)^\s*version:\s*"([^"]+)"' -Value $Version

        # Keep the frontend package version aligned with the desktop product version.
        $packagePath = Join-Path $RootDir 'frontend/package.json'
        if (Test-Path $packagePath) {
            Set-FirstGroupValue -Path $packagePath -Pattern '(?m)^\s*"version":\s*"([^"]+)"' -Value $Version
        }

        # Regenerate derived assets (exe resources, NSIS defines, manifest) from the config.
        $buildDir = Join-Path $RootDir 'build'
        $before = @(Get-ChildItem -Path $buildDir -Directory | ForEach-Object { $_.Name })
        Push-Location $buildDir
        try {
            & wails3 update build-assets -name $AppName -binaryname $AppName -config config.yml -dir . | Out-Null
            if ($LASTEXITCODE -ne 0) {
                throw 'wails3 update build-assets failed'
            }
        }
        finally {
            Pop-Location
        }

        # Drop platform asset dirs created by the generator to keep this Windows-only repo clean.
        $after = @(Get-ChildItem -Path $buildDir -Directory | ForEach-Object { $_.Name })
        foreach ($name in ($after | Where-Object { $before -notcontains $_ })) {
            Remove-Item -Path (Join-Path $buildDir $name) -Recurse -Force
            Write-Host "Removed generated asset dir build/$name"
        }
    }

    # Verify every asset carries the requested version so the artifacts cannot drift.
    Assert-VersionPattern -Path $configPath -Expected $Version `
        -Pattern ('(?m)^\s*version:\s*"' + [regex]::Escape($Version) + '"')
    Assert-VersionPattern -Path (Join-Path $RootDir 'build/windows/info.json') -Expected $Version `
        -Pattern ('"file_version":\s*"' + [regex]::Escape($Version) + '"')
    Assert-VersionPattern -Path (Join-Path $RootDir 'build/windows/info.json') -Expected $Version `
        -Pattern ('"ProductVersion":\s*"' + [regex]::Escape($Version) + '"')
    Assert-VersionPattern -Path (Join-Path $RootDir 'build/windows/nsis/wails_tools.nsh') -Expected $Version `
        -Pattern ('INFO_PRODUCTVERSION\s+"' + [regex]::Escape($Version) + '"')
    Assert-VersionPattern -Path (Join-Path $RootDir 'build/windows/wails.exe.manifest') -Expected $Version `
        -Pattern ('version="' + [regex]::Escape($Version) + '"')
    Assert-VersionPattern -Path (Join-Path $RootDir 'frontend/package.json') -Expected $Version `
        -Pattern ('"version":\s*"' + [regex]::Escape($Version) + '"') -Optional

    # Fail when the Go build no longer embeds the version, otherwise the in-app version drifts silently.
    $taskPath = Join-Path $RootDir 'build/windows/Taskfile.yml'
    if (-not ([IO.File]::ReadAllText($taskPath) -match 'internal/constant\.Version=\{\{\.VERSION')) {
        throw "Go version injection is missing in $taskPath"
    }

    Write-Host "Version $Version synced into build assets."
    return
}

New-Item -ItemType Directory -Path $releaseDir -Force | Out-Null

# Move the installer instead of copying to keep bin/ free of unversioned artifacts.
$installerSource = Join-Path $binDir "kairos-$Arch-installer.exe"
if (-not (Test-Path $installerSource)) {
    throw "Installer not found: $installerSource"
}
$installerTarget = Join-Path $releaseDir "kairos-$Version-$Arch-installer.exe"
Move-Item -Path $installerSource -Destination $installerTarget -Force

# Reject a stale installer so the release never ships a mismatched version.
$installerVersion = [System.Diagnostics.FileVersionInfo]::GetVersionInfo($installerTarget).ProductVersion
if ($installerVersion -ne $Version) {
    throw "Installer version '$installerVersion' does not match VERSION '$Version'"
}

$exeSource = Join-Path $binDir 'kairos.exe'
if (-not (Test-Path $exeSource)) {
    throw "Executable not found: $exeSource"
}
$configSource = Join-Path $RootDir 'configs/config.yaml'
if (-not (Test-Path $configSource)) {
    throw "Config file not found: $configSource"
}

# Stage a portable layout so the archive unpacks into a single self-contained folder.
$stageRoot = Join-Path $binDir 'release/.stage'
$stageDir = Join-Path $stageRoot 'Kairos'
if (Test-Path $stageRoot) {
    Remove-Item -Path $stageRoot -Recurse -Force
}
try {
    New-Item -ItemType Directory -Path (Join-Path $stageDir 'configs') -Force | Out-Null
    New-Item -ItemType Directory -Path (Join-Path $stageDir 'data') -Force | Out-Null
    Copy-Item -Path $exeSource -Destination (Join-Path $stageDir 'kairos.exe')
    Copy-Item -Path $configSource -Destination (Join-Path $stageDir 'configs/config.yaml')

    # Keep the data directory non-empty so archive round-trips preserve it and the app stays portable.
    New-Item -ItemType File -Path (Join-Path $stageDir 'data/.keep') | Out-Null

    $portableTarget = Join-Path $releaseDir "kairos-$Version-$Arch-portable.zip"
    if (Test-Path $portableTarget) {
        Remove-Item -Path $portableTarget -Force
    }
    Compress-Archive -Path $stageDir -DestinationPath $portableTarget -CompressionLevel Optimal
}
finally {
    # Remove the staging dir even when archiving fails.
    if (Test-Path $stageRoot) {
        Remove-Item -Path $stageRoot -Recurse -Force
    }
}

# Publish checksums so users can verify the downloaded artifacts.
$checksumPath = Join-Path $releaseDir 'SHA256SUMS.txt'
$lines = Get-ChildItem -Path $releaseDir -File |
    Where-Object { $_.Extension -in '.exe', '.zip' } |
    Sort-Object Name |
    ForEach-Object { "{0}  {1}" -f (Get-FileHash -Path $_.FullName -Algorithm SHA256).Hash.ToLower(), $_.Name }
Set-Content -Path $checksumPath -Value $lines -Encoding ascii

Write-Host "Release v$Version written to $releaseDir"
Get-ChildItem -Path $releaseDir -File | ForEach-Object {
    Write-Host ("  {0} ({1:N2} MB)" -f $_.Name, ($_.Length / 1MB))
}
