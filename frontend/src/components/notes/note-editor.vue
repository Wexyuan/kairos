<script setup lang="ts">
import { Crepe } from '@milkdown/crepe'
import { replaceAll } from '@milkdown/kit/utils'
import { basename } from 'pathe'
import { useAction } from '@/composables'
import { useNotesStore } from '@/stores/notes'
import '@milkdown/crepe/theme/common/style.css'
import '@milkdown/crepe/theme/classic.css'

const store = useNotesStore()
const { runAction } = useAction()
const message = useMessage()

const AUTOSAVE_DELAY = 800

const editorRef = ref<HTMLElement | null>(null)
const loading = ref(false)
const saving = ref(false)
const dirty = ref(false)
const wordCount = ref(0)

let crepe: Crepe | null = null // editor instance
let editorReady: Promise<Crepe> | null = null // guards against concurrent creation
let loadedPath = '' // path of the currently loaded note
let currentMarkdown = '' // current markdown content
let replacing = false // whether the editor is replacing existing content
let saveTimer: ReturnType<typeof setTimeout> | undefined // autosave timer
let saveTask: Promise<boolean> | null = null // in-flight save task

const noteTitle = computed(() => {
  const name = basename(store.activePath)
  return name ? name.replace(/\.md$/i, '') : '未选择笔记'
})

const saveStatus = computed(() => {
  if (loading.value)
    return '加载中'
  if (saving.value)
    return '保存中'
  return dirty.value ? '未保存' : '已保存'
})

const statusTextClass = computed(() => (loading.value || saving.value || dirty.value ? 'text-[#aa6622]' : 'text-ink/45'))

const statusDotClass = computed(() => {
  if (loading.value || saving.value)
    return 'bg-[#aa6622] animate-pulse'
  return dirty.value ? 'bg-[#aa6622]' : 'bg-[#227755]'
})

onMounted(() => {
  store.registerEditorFlush(flushEditor)
  window.addEventListener('keydown', handleKeydown)
  if (store.activePath)
    void loadNote(store.activePath)
})

onBeforeUnmount(() => {
  store.registerEditorFlush(null)
  window.removeEventListener('keydown', handleKeydown)
  clearTimeout(saveTimer)
  void saveNote()
  const instance = crepe
  crepe = null
  void instance?.destroy()
})

watch(() => store.activePath, (path) => {
  void handleNoteSwitch(path)
})

function handleKeydown(e: KeyboardEvent): void {
  if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 's' && store.activePath) {
    e.preventDefault()
    void flushEditor()
  }
}

async function selectWorkspace(): Promise<void> {
  await runAction('加载笔记', () => store.selectWorkspace())
}

async function createEditor(): Promise<Crepe> {
  const root = editorRef.value
  if (!root)
    throw new Error('编辑器容器未就绪')
  const instance = new Crepe({
    root,
    defaultValue: '',
    features: {
      [Crepe.Feature.TopBar]: true,
    },
    featureConfigs: {
      [Crepe.Feature.TopBar]: {
        headingOptions: [
          { label: '正文', level: null },
          { label: '标题 1', level: 1 },
          { label: '标题 2', level: 2 },
          { label: '标题 3', level: 3 },
          { label: '标题 4', level: 4 },
          { label: '标题 5', level: 5 },
          { label: '标题 6', level: 6 },
        ],
      },
      [Crepe.Feature.Placeholder]: { text: '开始书写…', mode: 'doc' },
      [Crepe.Feature.BlockEdit]: {
        textGroup: {
          label: '文本',
          text: { label: '正文' },
          h1: { label: '一级标题' },
          h2: { label: '二级标题' },
          h3: { label: '三级标题' },
          h4: { label: '四级标题' },
          h5: { label: '五级标题' },
          h6: { label: '六级标题' },
          quote: { label: '引用' },
          divider: { label: '分割线' },
        },
        listGroup: {
          label: '列表',
          bulletList: { label: '无序列表' },
          orderedList: { label: '有序列表' },
          taskList: { label: '任务列表' },
        },
        advancedGroup: {
          label: '高级',
          image: { label: '图片' },
          codeBlock: { label: '代码块' },
          table: { label: '表格' },
          math: { label: '公式' },
        },
      },
    },
  })
  instance.on((api) => {
    api.markdownUpdated((_ctx, markdown) => {
      currentMarkdown = markdown
      wordCount.value = countWords(markdown)
      if (replacing)
        return
      dirty.value = true
      scheduleAutosave()
    })
  })
  await instance.create()
  crepe = instance
  return instance
}

async function handleNoteSwitch(path: string): Promise<void> {
  clearTimeout(saveTimer)
  if (loadedPath && dirty.value)
    await saveNote()
  if (path)
    await loadNote(path)
}

async function loadNote(path: string): Promise<void> {
  loading.value = true
  try {
    try {
      const instance = await ensureEditor()
      const content = await store.readNote(path)
      if (crepe !== instance || store.activePath !== path)
        return
      replacing = true
      currentMarkdown = content
      wordCount.value = countWords(content)
      instance.editor.action(replaceAll(content))
      loadedPath = path
      dirty.value = false
    }
    catch (err) {
      message.error(`加载笔记失败：${err instanceof Error ? err.message : String(err)}`)
    }
  }
  finally {
    replacing = false
    loading.value = false
  }
}

function countWords(markdown: string): number {
  const cjk = markdown.match(/[\u3400-\u9FFF]/g)?.length ?? 0
  const latin = markdown.match(/[a-z0-9]+/gi)?.length ?? 0
  return cjk + latin
}

function saveNote(): Promise<boolean> {
  if (saveTask)
    return saveTask
  if (!loadedPath || !dirty.value)
    return Promise.resolve(true)
  saveTask = (async () => {
    saving.value = true
    try {
      const ok = await runAction('保存', () => store.saveNote(loadedPath, currentMarkdown))
      if (ok)
        dirty.value = false
      return ok
    }
    finally {
      saving.value = false
      saveTask = null
    }
  })()
  return saveTask
}

function scheduleAutosave(): void {
  clearTimeout(saveTimer)
  saveTimer = setTimeout(() => void saveNote(), AUTOSAVE_DELAY)
}

async function flushEditor(): Promise<boolean> {
  clearTimeout(saveTimer)
  return await saveNote()
}

function ensureEditor(): Promise<Crepe> {
  if (!editorReady) {
    editorReady = createEditor().catch((err) => {
      editorReady = null
      throw err
    })
  }
  return editorReady
}
</script>

<template>
  <div class="panel flex flex-col h-full">
    <div class="px-4 py-2.5 border-b border-ink/5 flex shrink-0 gap-2 items-center">
      <span class="i-mdi-file-document-outline text-ink/45 shrink-0 h-4 w-4" />
      <div class="flex flex-1 min-h-8 min-w-0 items-center" :title="store.activePath || noteTitle">
        <span class="title-sm truncate">
          {{ noteTitle }}
        </span>
      </div>
      <template v-if="store.activePath">
        <span class="text-xs text-ink/45 shrink-0">{{ wordCount }} 字</span>
        <span class="text-xs flex-center shrink-0 gap-1" :class="statusTextClass">
          <span class="rounded-full h-1.5 w-1.5" :class="statusDotClass" />
          {{ saveStatus }}
        </span>
        <n-button quaternary size="small" class="icon-btn shrink-0" title="保存（Ctrl+S）" :disabled="!dirty" @click="flushEditor">
          <template #icon>
            <span class="i-mdi-content-save-outline h-4 w-4" />
          </template>
        </n-button>
      </template>
    </div>

    <div v-if="!store.root" class="flex-center flex-1">
      <n-empty description="选择工作区开始记录叭 (^v^)">
        <template #extra>
          <n-button size="small" type="primary" @click="selectWorkspace">
            选择工作区
          </n-button>
        </template>
      </n-empty>
    </div>
    <div v-else-if="!store.activePath" class="flex-center flex-1">
      <n-empty description="选择或新建一篇笔记叭 (^v^)" />
    </div>
    <div v-show="store.activePath" class="note-body flex-1 min-h-0 overflow-y-auto">
      <div ref="editorRef" class="h-full" />
    </div>
  </div>
</template>

<style scoped>
:deep(.milkdown) {
  --crepe-color-background: var(--app-page);
  --crepe-color-on-background: var(--app-ink);
  --crepe-color-surface: var(--app-surface);
  --crepe-color-surface-low: var(--app-page);
  --crepe-color-on-surface: var(--app-ink);
  --crepe-color-on-surface-variant: color-mix(in srgb, var(--app-ink) 80%, transparent);
  --crepe-color-outline: color-mix(in srgb, var(--app-ink) 45%, transparent);
  --crepe-color-primary: var(--app-brand);
  --crepe-color-secondary: color-mix(in srgb, var(--app-brand) 15%, var(--app-page));
  --crepe-color-on-secondary: var(--app-ink);
  --crepe-color-inverse: var(--app-ink);
  --crepe-color-on-inverse: var(--app-page);
  --crepe-color-hover: color-mix(in srgb, var(--app-brand) 10%, transparent);
  --crepe-color-selected: color-mix(in srgb, var(--app-brand) 15%, transparent);
  --crepe-color-inline-area: color-mix(in srgb, var(--app-ink) 10%, transparent);
  --crepe-color-inline-code: color-mix(in srgb, var(--app-brand) 80%, var(--app-ink));
  --crepe-color-error: #d03050;

  --crepe-base-font-size: 15px;
  --crepe-font-default: var(--app-font-sans);
  --crepe-font-title: var(--app-font-sans);
  --crepe-font-code: var(--app-font-mono);

  --crepe-shadow-1: 0 2px 8px rgb(24 34 46 / 6%), 0 0 0 1px color-mix(in srgb, var(--app-ink) 10%, transparent);
  --crepe-shadow-2: 0 8px 24px rgb(24 34 46 / 8%), 0 0 0 1px color-mix(in srgb, var(--app-ink) 10%, transparent);
}

:deep(.milkdown .ProseMirror) {
  max-width: 800px;
  margin: 0 auto;
  padding: 32px 72px 56px;
}

:deep(.milkdown .milkdown-top-bar) {
  padding: 10px 16px;
  border-bottom: 1px solid color-mix(in srgb, var(--app-ink) 5%, transparent);
  background: var(--app-surface);
}

:deep(.milkdown .milkdown-top-bar .top-bar-inner) {
  justify-content: center;
  gap: 2px;
  max-width: 800px;
  margin: 0 auto;
}

:deep(.milkdown .milkdown-top-bar .top-bar-heading-selector) {
  padding: 0;
}

:deep(.milkdown .milkdown-top-bar .top-bar-item) {
  width: 32px;
  height: 32px;
  margin: 0;
  padding: 0;
  border-radius: 8px;
  color: color-mix(in srgb, var(--app-ink) 60%, transparent);
}

:deep(.milkdown .milkdown-top-bar .top-bar-item svg) {
  width: 16px;
  height: 16px;
  fill: currentColor;
}

:deep(.milkdown .milkdown-top-bar .top-bar-item.active) {
  background: color-mix(in srgb, var(--app-brand) 10%, transparent);
}

:deep(.milkdown .milkdown-top-bar .top-bar-item:hover svg) {
  color: var(--app-brand);
  fill: var(--app-brand);
}

:deep(.milkdown .milkdown-top-bar .top-bar-divider) {
  height: 20px;
  margin: 0 4px;
  background: color-mix(in srgb, var(--app-ink) 10%, transparent);
}

:deep(.milkdown .milkdown-top-bar .top-bar-heading-button) {
  height: 32px;
  border-radius: 8px;
}

:deep(.milkdown .milkdown-top-bar .top-bar-heading-button .top-bar-heading-label) {
  min-width: 48px;
  font-size: 14px;
  font-weight: 600;
}

:deep(.milkdown .milkdown-top-bar .top-bar-heading-dropdown) {
  padding: 8px;
  border-radius: 12px;
}

:deep(.milkdown .milkdown-top-bar .top-bar-heading-option) {
  border-radius: 8px;
  font-size: 14px;
}

:deep(.milkdown .milkdown-toolbar) {
  align-items: center;
  gap: 2px;
  padding: 4px;
  border-radius: 12px;
}

:deep(.milkdown .milkdown-toolbar .divider) {
  height: 20px;
  margin: 0 4px;
  background: color-mix(in srgb, var(--app-ink) 10%, transparent);
}

:deep(.milkdown .milkdown-toolbar .toolbar-item) {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  margin: 0;
  padding: 0;
  border-radius: 8px;
  color: color-mix(in srgb, var(--app-ink) 60%, transparent);
}

:deep(.milkdown .milkdown-toolbar .toolbar-item svg) {
  width: 16px;
  height: 16px;
  fill: currentColor;
}

:deep(.milkdown .milkdown-toolbar .toolbar-item.active) {
  background: color-mix(in srgb, var(--app-brand) 10%, transparent);
}

:deep(.milkdown .milkdown-toolbar .toolbar-item:hover svg) {
  color: var(--app-brand);
  fill: var(--app-brand);
}

:deep(.milkdown .milkdown-block-handle) {
  padding: 4px 0;
}

:deep(.milkdown .milkdown-block-handle .operation-item) {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  padding: 0;
  border-radius: 8px;
}

:deep(.milkdown .milkdown-block-handle .operation-item svg) {
  width: 16px;
  height: 16px;
}

.note-body {
  scrollbar-width: thin;
  scrollbar-color: color-mix(in srgb, var(--app-brand) 20%, transparent) transparent;
}

.note-body::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}

.note-body::-webkit-scrollbar-thumb {
  border-radius: 999px;
  background: color-mix(in srgb, var(--app-brand) 20%, transparent);
}

.note-body::-webkit-scrollbar-thumb:hover {
  background: color-mix(in srgb, var(--app-brand) 40%, transparent);
}

:deep(.milkdown .ProseMirror > *:first-child) {
  margin-top: 0;
}

:deep(.milkdown .ProseMirror p) {
  padding: 3px 0;
  line-height: 1.7;
}

:deep(.milkdown .ProseMirror h1),
:deep(.milkdown .ProseMirror h2),
:deep(.milkdown .ProseMirror h3),
:deep(.milkdown .ProseMirror h4),
:deep(.milkdown .ProseMirror h5),
:deep(.milkdown .ProseMirror h6) {
  padding: 0;
  font-weight: 600;
  letter-spacing: -0.01em;
  color: var(--app-ink);
}

:deep(.milkdown .ProseMirror h1) {
  margin: 24px 0 8px;
  font-size: 28px;
  line-height: 1.4;
}

:deep(.milkdown .ProseMirror h2) {
  margin: 20px 0 6px;
  font-size: 24px;
  line-height: 1.45;
}

:deep(.milkdown .ProseMirror h3) {
  margin: 18px 0 6px;
  font-size: 20px;
  line-height: 1.5;
}

:deep(.milkdown .ProseMirror h4) {
  margin: 16px 0 4px;
  font-size: 18px;
  line-height: 1.55;
}

:deep(.milkdown .ProseMirror h5) {
  margin: 14px 0 4px;
  font-size: 16px;
  line-height: 1.6;
}

:deep(.milkdown .ProseMirror h6) {
  margin: 12px 0 4px;
  font-size: 15px;
  line-height: 1.6;
  color: color-mix(in srgb, var(--app-ink) 60%, transparent);
}

:deep(.milkdown .ProseMirror *::selection) {
  background: color-mix(in srgb, var(--app-brand) 15%, transparent);
}

:deep(.milkdown .ProseMirror a) {
  text-decoration-color: color-mix(in srgb, var(--app-brand) 40%, transparent);
  text-underline-offset: 2px;
}

:deep(.milkdown .ProseMirror a:hover) {
  text-decoration-color: var(--app-brand);
}

:deep(.milkdown .ProseMirror code) {
  padding: 1px 5px;
  border-radius: 4px;
}

:deep(.milkdown .ProseMirror blockquote) {
  margin: 8px 0;
  padding-left: 16px;
}

:deep(.milkdown .ProseMirror blockquote::before) {
  background: color-mix(in srgb, var(--app-brand) 40%, transparent);
}

:deep(.milkdown .ProseMirror blockquote p) {
  color: color-mix(in srgb, var(--app-ink) 80%, transparent);
}

:deep(.milkdown .milkdown-table-block th),
:deep(.milkdown .milkdown-table-block td) {
  padding: 6px 14px;
  border-color: color-mix(in srgb, var(--app-ink) 10%, transparent);
}

:deep(.milkdown .milkdown-table-block th) {
  font-weight: 600;
  background: color-mix(in srgb, var(--app-ink) 5%, transparent);
}

:deep(.milkdown .milkdown-list-item-block .label.checked),
:deep(.milkdown .milkdown-list-item-block .label.unchecked) {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 32px;
  padding: 0;
}

:deep(.milkdown .milkdown-list-item-block .label.checked svg),
:deep(.milkdown .milkdown-list-item-block .label.unchecked svg) {
  width: 16px;
  height: 16px;
}

:deep(.milkdown .milkdown-list-item-block .label.checked svg) {
  fill: var(--app-brand);
}

:deep(.milkdown .milkdown-list-item-block .label.checked:hover svg) {
  fill: color-mix(in srgb, var(--app-brand) 80%, var(--app-ink));
}

:deep(.milkdown .milkdown-list-item-block .label.unchecked svg) {
  fill: color-mix(in srgb, var(--app-ink) 40%, transparent);
}

:deep(.milkdown .milkdown-list-item-block .label.unchecked:hover svg) {
  fill: color-mix(in srgb, var(--app-ink) 60%, transparent);
}

:deep(.milkdown .milkdown-code-block) {
  margin: 12px 0;
  padding: 8px 16px 16px;
  border-radius: 8px;
  background: #282c34;
}

:deep(.milkdown .milkdown-code-block .cm-editor) {
  background: transparent;
}

:deep(.milkdown .milkdown-code-block .cm-content),
:deep(.milkdown .milkdown-code-block .cm-gutters) {
  font-family: var(--crepe-font-code);
  font-size: 14px;
  background: transparent;
}

:deep(.milkdown .milkdown-code-block .cm-line) {
  line-height: 1.7;
}

:deep(.milkdown .milkdown-code-block .milkdown-code-block-placeholder) {
  font-size: 14px;
  color: rgb(255 255 255 / 45%);
}

:deep(.milkdown .milkdown-code-block .tools .language-button) {
  border-radius: 4px;
  background: rgb(255 255 255 / 10%);
  color: rgb(255 255 255 / 70%);
}

:deep(.milkdown .milkdown-code-block .tools .language-button:hover) {
  background: rgb(255 255 255 / 15%);
}

:deep(.milkdown .milkdown-code-block .tools .language-button .expand-icon svg) {
  color: rgb(255 255 255 / 55%);
}

:deep(.milkdown .milkdown-code-block .tools .tools-button-group button) {
  border-radius: 4px;
  background: rgb(255 255 255 / 10%);
  color: rgb(255 255 255 / 75%);
}

:deep(.milkdown .milkdown-code-block .tools .tools-button-group button svg) {
  fill: rgb(255 255 255 / 75%);
}

:deep(.milkdown .milkdown-code-block .tools .tools-button-group button:hover) {
  background: rgb(255 255 255 / 20%);
}
</style>
