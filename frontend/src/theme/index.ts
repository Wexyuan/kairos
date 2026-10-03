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
  const surface = mix(page, brand, 0.05)
  const brandHover = mix(brand, '#ffffff', 0.12)
  const brandPressed = mix(brand, '#000000', 0.12)

  return {
    common: {
      borderRadius: '8px',
      borderRadiusSmall: '4px',
      primaryColor: brand,
      primaryColorHover: brandHover,
      primaryColorPressed: brandPressed,
      primaryColorSuppl: brand,
      infoColor: brand,
      infoColorHover: brandHover,
      infoColorPressed: brandPressed,
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
      dividerColor: hexToRgba(ink, 0.15),
      borderColor: hexToRgba(ink, 0.15),
      closeIconColor: hexToRgba(ink, 0.6),
      closeIconColorHover: hexToRgba(ink, 0.75),
      closeIconColorPressed: ink,
      closeColorHover: hexToRgba(ink, 0.1),
      closeColorPressed: hexToRgba(ink, 0.15),
      clearColor: hexToRgba(ink, 0.6),
      clearColorHover: hexToRgba(ink, 0.75),
      clearColorPressed: ink,
      scrollbarColor: hexToRgba(ink, 0.1),
      scrollbarColorHover: hexToRgba(ink, 0.3),
      scrollbarWidth: '4px',
      scrollbarHeight: '4px',
      scrollbarBorderRadius: '4px',
      progressRailColor: hexToRgba(ink, 0.15),
      railColor: hexToRgba(ink, 0.15),
      cardColor: surface,
      modalColor: mix(page, ink, 0.15),
      bodyColor: page,
      inputColor: 'transparent',
      actionColor: hexToRgba(ink, 0.1),
      hoverColor: hexToRgba(ink, 0.1),
      pressedColor: hexToRgba(ink, 0.15),
      inputColorDisabled: hexToRgba(ink, 0.1),
      buttonColor2: hexToRgba(ink, 0.05),
      buttonColor2Hover: hexToRgba(ink, 0.1),
      buttonColor2Pressed: hexToRgba(ink, 0.15),
    },
    Layout: {
      siderColor: page,
      siderToggleButtonBorder: '1px solid transparent',
      siderToggleButtonColor: hexToRgba(brand, 0.1),
      siderToggleButtonIconColor: hexToRgba(brand, 0.6),
    },
    Menu: {
      itemColorHover: hexToRgba(brand, 0.1),
      itemHeight: '36px',
    },
    Radio: {
      buttonColor: 'transparent',
      buttonColorActive: hexToRgba(brand, 0.8),
      buttonTextColorActive: page,
      buttonBoxShadowFocus: 'inset 0 0 0 1px #0000',
    },
  }
}
