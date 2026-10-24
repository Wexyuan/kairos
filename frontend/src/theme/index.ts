import type { GlobalThemeOverrides } from 'naive-ui'

// hexToRgb converts a #RGB or #RRGGBB string to an array of RGB values.
function hexToRgb(hex: string): [number, number, number] {
  const digits = hex.trim().replace(/^#/, '')
  const full = digits.length === 3 ? digits.replace(/./g, c => c + c) : digits
  return [0, 2, 4].map(i => Number.parseInt(full.slice(i, i + 2), 16)) as [number, number, number]
}

// hexToRgba converts a #RGGBB string to an rgba string.
function hexToRgba(hex: string, opacity: number): string {
  const [r, g, b] = hexToRgb(hex)
  return `rgba(${r}, ${g}, ${b}, ${opacity})`
}

// mix mixes a hex color towards another hex color by a ratio.
function mix(hex: string, toward: string, ratio: number): string {
  const [r1, g1, b1] = hexToRgb(hex)
  const [r2, g2, b2] = hexToRgb(toward)
  const channel = (a: number, b: number) => Math.round(a + (b - a) * ratio).toString(16).padStart(2, '0')
  return `#${channel(r1, r2)}${channel(g1, g2)}${channel(b1, b2)}`
}

// getThemeVars retrieves the theme variables.
function getThemeVars(dark: boolean): { ink: string, page: string, brand: string } {
  const probe = document.createElement('div')
  probe.className = dark ? 'theme-dark' : 'theme-light'
  probe.style.display = 'none'
  document.body.appendChild(probe)
  const style = getComputedStyle(probe)
  const vars = {
    ink: style.getPropertyValue('--app-ink').trim(),
    page: style.getPropertyValue('--app-page').trim(),
    brand: style.getPropertyValue('--app-brand').trim(),
  }
  probe.remove()
  return vars
}

// buildThemeOverrides builds naive-ui theme overrides.
export function buildThemeOverrides(dark: boolean): GlobalThemeOverrides {
  const { ink, page, brand } = getThemeVars(dark)

  return {
    common: {
      fontFamily: 'var(--app-font-sans)',
      fontFamilyMono: 'var(--app-font-mono)',
      borderRadius: '12px',
      borderRadiusSmall: '8px',
      primaryColor: brand,
      primaryColorHover: mix(brand, '#ffffff', 0.12),
      primaryColorPressed: mix(brand, '#000000', 0.12),
      primaryColorSuppl: brand,
      infoColor: brand,
      infoColorHover: mix(brand, '#ffffff', 0.12),
      infoColorPressed: mix(brand, '#000000', 0.12),
      infoColorSuppl: brand,
      textColorBase: ink,
      textColor1: ink,
      textColor2: hexToRgba(ink, 0.8),
      textColor3: hexToRgba(ink, 0.6),
      textColorDisabled: hexToRgba(ink, 0.3),
      placeholderColor: hexToRgba(ink, 0.45),
      placeholderColorDisabled: hexToRgba(ink, 0.3),
      iconColor: hexToRgba(ink, 0.6),
      iconColorHover: hexToRgba(ink, 0.75),
      iconColorPressed: ink,
      iconColorDisabled: hexToRgba(ink, 0.3),
      dividerColor: hexToRgba(ink, 0.05),
      borderColor: 'transparent',
      closeIconColor: hexToRgba(ink, 0.6),
      closeIconColorHover: hexToRgba(ink, 0.75),
      closeIconColorPressed: ink,
      closeColorHover: hexToRgba(ink, 0.1),
      closeColorPressed: hexToRgba(ink, 0.15),
      clearColor: hexToRgba(ink, 0.6),
      clearColorHover: hexToRgba(ink, 0.75),
      clearColorPressed: ink,
      scrollbarColor: hexToRgba(brand, 0.20),
      scrollbarColorHover: hexToRgba(brand, 0.5),
      scrollbarWidth: '6px',
      scrollbarHeight: '6px',
      scrollbarBorderRadius: '999px',
      popoverColor: page,
      cardColor: page,
      modalColor: page,
      bodyColor: page,
      inputColor: mix(page, ink, dark ? 0.07 : 0.12),
      hoverColor: hexToRgba(ink, 0.05),
      pressedColor: hexToRgba(ink, 0.07),
      inputColorDisabled: mix(page, ink, dark ? 0.07 : 0.12),
      boxShadow1: '0 2px 8px rgba(24,34,46,.06)',
      boxShadow2: '0 8px 24px rgba(24,34,46,.08)',
      boxShadow3: '0 16px 40px rgba(24,34,46,.12)',
    },
    Layout: {
      siderToggleButtonBorder: '1px solid transparent',
      siderToggleButtonColor: hexToRgba(brand, 0.1),
      siderToggleButtonIconColor: hexToRgba(brand, 0.6),
    },
    Menu: {
      itemIconColor: hexToRgba(ink, 0.8),
      itemHeight: '36px',
    },
    Button: {
      heightTiny: '28px',
      heightSmall: '32px',
      heightMedium: '40px',
      borderRadiusTiny: '8px',
      borderRadiusSmall: '8px',
      fontWeight: '500',
      iconSizeTiny: '16px',
      iconSizeSmall: '16px',
    },
    Input: {
      heightSmall: '32px',
      heightMedium: '40px',
      textColor: ink,
      colorFocus: page,
      borderHover: '1px solid transparent',
      borderFocus: '1px solid transparent',
      boxShadowFocus: `0 0 0 3px ${hexToRgba(brand, 0.12)}`,
    },
    Tree: {
      nodeHeight: '36px',
      arrowColor: hexToRgba(ink, 0.45),
    },
    Switch: {
      railColor: mix(page, ink, dark ? 0.07 : 0.12),
      buttonColor: page,
      boxShadowFocus: `0 0 0 3px ${hexToRgba(brand, 0.12)}`,
    },
    Empty: {
      textColor: hexToRgba(ink, 0.6),
      iconColor: hexToRgba(ink, 0.3),
    },
    Dialog: {
      border: 'none',
      borderRadius: '16px',
      padding: '24px',
    },
    Dropdown: {
      optionHeightMedium: '36px',
      optionTextColorHover: ink,
      optionColorHover: hexToRgba(brand, 0.1),
    },
  }
}
