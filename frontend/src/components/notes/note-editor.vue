<script setup lang="ts">
import type { Node as ProseMirrorNode, Schema } from '@milkdown/kit/prose/model'
import { Crepe } from '@milkdown/crepe'
import { uploadConfig } from '@milkdown/kit/plugin/upload'
import { $remark, replaceAll } from '@milkdown/kit/utils'
import { basename } from 'pathe'
import { useAction } from '@/composables'
import { useAppStore } from '@/stores/app'
import { useNotesStore } from '@/stores/notes'
import { createNoteImageController } from './note-image'
import '@milkdown/crepe/theme/common/style.css'
import '@milkdown/crepe/theme/classic.css'

const store = useNotesStore()
const appStore = useAppStore()
const { runAction } = useAction()
const message = useMessage()

const NOTE_IMAGE_EXTENSION = /\.(?:png|jpe?g|gif|webp)$/i
const AUTOSAVE_DELAY = 800

const editorRef = ref<HTMLElement | null>(null)
const loading = ref(false)
const saving = ref(false)
const dirty = ref(false)
const wordCount = ref(0)

let crepe: Crepe | null = null // editor instance
let editorReady: Promise<Crepe> | null = null // guards against concurrent creation
let loadedRoot = '' // workspace root of the currently loaded note
let loadedPath = '' // path of the currently loaded note
let currentMarkdown = '' // current markdown content
let lastSavedMarkdown = '' // markdown of the last successful load or save
let saveTimer: ReturnType<typeof setTimeout> | undefined // autosave timer
let saveTask: Promise<boolean> | null = null // in-flight save task
let noteSwitchRequest = 0 // only the latest note switch may replace editor content

const noteMaxWidthCss = computed(() => `${appStore.noteMaxWidth}px`)

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

const normalizeNoteImages = $remark('normalize-note-images', () => () => (tree) => {
  const stack = [tree as NoteImageMarkdownNode]
  while (stack.length > 0) {
    const node = stack.pop()
    if (!node)
      continue
    if (node.type === 'image' || node.type === 'image-block') {
      node.title ??= ''
      node.alt ??= ''
    }
    if (node.children)
      stack.push(...node.children)
  }
})

const imagePasteHandled = Symbol('note-image-paste')

interface NoteImagePasteEvent extends ClipboardEvent {
  [imagePasteHandled]?: boolean
}

interface NoteImageMarkdownNode {
  type?: string
  title?: string | null
  alt?: string | null
  children?: NoteImageMarkdownNode[]
}

const noteImageController = createNoteImageController({
  getContext: () => ({ root: loadedRoot, notePath: loadedPath }),
  getNoteMaxImageSize: () => appStore.noteMaxImageSize,
  saveNoteImage: (root, notePath, data) => store.saveNoteImage(root, notePath, data),
  readNoteImage: (root, notePath, path) => store.readNoteImage(root, notePath, path),
  reportError: (error) => {
    message.error(error)
  },
})

onMounted(() => {
  store.registerEditorFlush(flushEditor)
  window.addEventListener('keydown', handleKeydown)
  editorRef.value?.addEventListener('paste', handleImagePaste, true)
  if (store.activePath)
    void handleNoteSwitch(store.activePath)
})

onBeforeUnmount(() => {
  store.registerEditorFlush(null)
  window.removeEventListener('keydown', handleKeydown)
  editorRef.value?.removeEventListener('paste', handleImagePaste, true)
  clearTimeout(saveTimer)
  noteSwitchRequest++
  noteImageController.clearCache()
  const instance = crepe
  // keep the instance alive until the final save finishes, then destroy it.
  let timeoutId: ReturnType<typeof setTimeout> | undefined
  const timeout = new Promise<void>((resolve) => {
    timeoutId = setTimeout(resolve, 3000)
  })
  void Promise.race([flushEditor(), timeout])
    .catch(() => {})
    .finally(() => {
      clearTimeout(timeoutId)
      crepe = null
      void instance?.destroy()
    })
})

watch(() => store.activePath, (path) => {
  if (path === loadedPath)
    return
  void handleNoteSwitch(path)
})

function handleKeydown(e: KeyboardEvent): void {
  if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 's' && store.activePath) {
    e.preventDefault()
    void flushEditor()
  }
}

function handleImagePaste(event: ClipboardEvent): void {
  const pasteEvent = event as NoteImagePasteEvent
  if (pasteEvent[imagePasteHandled])
    return
  const clipboard = event.clipboardData
  const files = [...(clipboard?.files ?? [])].filter(isNoteImageFile)
  if (files.length === 0)
    return
  const target = event.target
  if (!(target instanceof Element) || !target.closest('.ProseMirror'))
    return
  if (target instanceof HTMLInputElement || target instanceof HTMLTextAreaElement)
    return
  if (target.closest('.cm-editor'))
    return

  const html = stripHtmlImages(clipboard?.getData('text/html') ?? '')
  const text = clipboard?.getData('text/plain') ?? ''

  let imageEvent: ClipboardEvent
  let textEvent: ClipboardEvent | null = null
  try {
    const imageData = new DataTransfer()
    for (const file of files)
      imageData.items.add(file)
    imageEvent = new ClipboardEvent('paste', {
      bubbles: true,
      cancelable: true,
      clipboardData: imageData,
    })
    // clipboardData cannot be set on every platform; warn instead of dropping silently.
    if (!imageEvent.clipboardData || imageEvent.clipboardData.files.length === 0) {
      message.warning('当前环境下不支持粘贴图片，请使用上传图片按钮')
      return
    }

    if (html.trim() || text.trim()) {
      const textData = new DataTransfer()
      if (html)
        textData.setData('text/html', html)
      textData.setData('text/plain', text)
      textEvent = new ClipboardEvent('paste', {
        bubbles: true,
        cancelable: true,
        clipboardData: textData,
      })
    }
  }
  catch {
    return
  }

  event.preventDefault()
  event.stopPropagation()
  const nextImageEvent = imageEvent as NoteImagePasteEvent
  nextImageEvent[imagePasteHandled] = true
  if (textEvent)
    target.dispatchEvent(textEvent)
  target.dispatchEvent(nextImageEvent)
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
        buildTopBar: (builder) => {
          const groups = builder.build()
          const groupByKey = new Map(groups.map(group => [group.key, group]))
          const textGroup = groupByKey.get('heading')
          const listGroup = groupByKey.get('list')
          const advancedGroup = groupByKey.get('insert')
          if (!textGroup || !listGroup || !advancedGroup)
            return

          textGroup.label = '文本'
          listGroup.label = '列表'
          advancedGroup.label = '高级'

          const formattingGroup = groupByKey.get('formatting')
          const quoteGroup = groupByKey.get('more')
          const blockGroup = groupByKey.get('block')
          if (formattingGroup)
            textGroup.items.push(...formattingGroup.items)
          if (quoteGroup)
            textGroup.items.push(...quoteGroup.items)
          if (blockGroup)
            advancedGroup.items.push(...blockGroup.items)

          const linkIndex = advancedGroup.items.findIndex(item => item.key === 'link')
          if (linkIndex >= 0)
            textGroup.items.push(...advancedGroup.items.splice(linkIndex, 1))

          const advancedOrder = ['image', 'code-block', 'table', 'math']
          const orderOf = (key: string) => {
            const index = advancedOrder.indexOf(key)
            return index < 0 ? advancedOrder.length : index
          }
          advancedGroup.items.sort((a, b) => orderOf(a.key) - orderOf(b.key))

          groups.splice(0, groups.length, textGroup, listGroup, advancedGroup)
        },
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
      [Crepe.Feature.ImageBlock]: noteImageController.config,
    },
  })
  instance.editor.use(normalizeNoteImages)
  instance.editor.config((ctx) => {
    ctx.update(uploadConfig.key, prev => ({
      ...prev,
      uploader: createImageUploader(),
    }))
  })
  instance.on((api) => {
    api.markdownUpdated(() => {
      const latest = crepe?.getMarkdown() ?? ''
      currentMarkdown = latest
      wordCount.value = countWords(latest)
      dirty.value = latest !== lastSavedMarkdown
      if (dirty.value)
        scheduleAutosave()
    })
  })
  await instance.create()
  crepe = instance
  return instance
}

function createImageUploader() {
  return async (files: FileList, schema: Schema): Promise<ProseMirrorNode[]> => {
    const nodeType = schema.nodes['image-block'] ?? schema.nodes.image
    if (!nodeType)
      return []
    const nodes: ProseMirrorNode[] = []
    for (const file of Array.from(files)) {
      if (!isNoteImageFile(file))
        continue
      try {
        const src = await noteImageController.config.onUpload(file)
        const node = nodeType.createAndFill({ src })
        if (node)
          nodes.push(node)
      }
      catch {
        // onUpload already reported the failure; skip the file and clear its placeholder.
      }
    }
    return nodes
  }
}

async function handleNoteSwitch(path: string): Promise<void> {
  const request = ++noteSwitchRequest
  clearTimeout(saveTimer)
  try {
    if (!await flushEditor()) {
      if (loadedPath)
        store.activePath = loadedPath
      return
    }
    if (request !== noteSwitchRequest)
      return
    noteImageController.clearCache()
    if (path)
      await loadNote(path, request)
  }
  finally {
    if (request === noteSwitchRequest)
      loading.value = false
  }
}

async function loadNote(path: string, request = noteSwitchRequest): Promise<void> {
  const previousRoot = loadedRoot
  const previousPath = loadedPath
  const previousMarkdown = currentMarkdown
  const previousLastSaved = lastSavedMarkdown
  const previousWordCount = wordCount.value
  const previousDirty = dirty.value
  loading.value = true
  try {
    const instance = await ensureEditor()
    const content = await store.readNote(path, store.root)
    if (request !== noteSwitchRequest || crepe !== instance || store.activePath !== path)
      return
    loadedRoot = store.root
    loadedPath = path
    instance.editor.action(replaceAll(content, true))
    lastSavedMarkdown = instance.getMarkdown()
    currentMarkdown = lastSavedMarkdown
    wordCount.value = countWords(lastSavedMarkdown)
    dirty.value = false
  }
  catch (err) {
    loadedRoot = previousRoot
    loadedPath = previousPath
    currentMarkdown = previousMarkdown
    lastSavedMarkdown = previousLastSaved
    wordCount.value = previousWordCount
    dirty.value = previousDirty
    if (request === noteSwitchRequest && store.activePath !== previousPath)
      store.activePath = previousPath
    message.error(`加载笔记失败：${err instanceof Error ? err.message : String(err)}`)
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
  if (!loadedPath || !crepe)
    return Promise.resolve(true)
  const path = loadedPath
  const content = crepe.getMarkdown()
  if (content === lastSavedMarkdown) {
    currentMarkdown = content
    dirty.value = false
    return Promise.resolve(true)
  }
  saveTask = (async () => {
    saving.value = true
    try {
      const ok = await runAction('保存', () => store.saveNote(path, content, loadedRoot))
      if (ok && loadedPath === path) {
        currentMarkdown = content
        lastSavedMarkdown = content
        dirty.value = crepe ? crepe.getMarkdown() !== content : false
        if (dirty.value)
          scheduleAutosave()
      }
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
  saveTimer = setTimeout(() => void runAutosave(), AUTOSAVE_DELAY)
}

async function runAutosave(): Promise<void> {
  if (saveTask)
    await saveTask
  await saveNote()
}

async function flushEditor(): Promise<boolean> {
  clearTimeout(saveTimer)
  for (let attempt = 0; attempt < 3; attempt++) {
    await noteImageController.waitForUploads()
    const saved = await saveNote()
    if (!saved)
      return false
    if (!noteImageController.hasPendingUploads() && !dirty.value)
      return true
  }
  return false
}

function isNoteImageFile(file: File): boolean {
  return file.type.startsWith('image/') || NOTE_IMAGE_EXTENSION.test(file.name)
}

function stripHtmlImages(html: string): string {
  if (!html)
    return ''
  const parsed = new DOMParser().parseFromString(html, 'text/html')
  parsed.querySelectorAll('img').forEach(image => image.remove())
  return parsed.body.innerHTML
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
        <n-button quaternary size="small" class="icon-btn shrink-0" title="保存（Ctrl+S）" :disabled="saving" @click="flushEditor">
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

  --note-max-width: v-bind(noteMaxWidthCss);
}

:deep(.milkdown .ProseMirror) {
  max-width: var(--note-max-width);
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
  max-width: var(--note-max-width);
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
