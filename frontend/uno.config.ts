import {
  defineConfig,
  presetAttributify,
  presetIcons,
  presetTypography,
  presetWind4,
  transformerDirectives,
  transformerVariantGroup,
} from 'unocss'

export default defineConfig({
  presets: [
    presetAttributify(),
    presetIcons({
      scale: 1.2,
      extraProperties: {
        'display': 'inline-block',
        'vertical-align': 'middle',
      },
      collections: {
        mdi: () => import('@iconify-json/mdi').then(i => i.icons),
      },
    }),
    presetTypography(),
    presetWind4(),
  ],
  transformers: [
    transformerDirectives(),
    transformerVariantGroup(),
  ],
  theme: {
    colors: {
      ink: 'var(--app-ink)',
      page: 'var(--app-page)',
      brand: 'var(--app-brand)',
      surface: 'var(--app-surface)',
    },
    shadow: {
      1: '0 2px 8px rgba(24,34,46,.06)',
      2: '0 8px 24px rgba(24,34,46,.08)',
      3: '0 16px 40px rgba(24,34,46,.12)',
    },
  },
  shortcuts: {
    'wh-full': 'w-full h-full',
    'wh-screen': 'w-screen h-screen',
    'flex-center': 'flex items-center justify-center',
    'title-lg': 'text-lg font-semibold text-ink',
    'title-md': 'text-base font-medium text-ink',
    'title-sm': 'text-sm font-medium text-ink',
    'text-body': 'text-base text-ink/60',
    'text-caption': 'text-sm text-ink/45',
    'icon-btn': '!p-0 !h-7 !w-7',
    'panel': 'bg-page rounded-2xl shadow-1 overflow-hidden',
  },
  preflights: [
    {
      getCSS: () => '*,::before,::after{border-color:color-mix(in srgb, var(--app-ink) 15%, transparent)}',
    },
  ],
  content: {
    pipeline: {
      include: [
        /\.(vue|svelte|[jt]sx|mdx?|astro|elm|php|phoenix|html|ts|js)($|\?)/,
      ],
      exclude: [/node_modules/],
    },
  },
})
