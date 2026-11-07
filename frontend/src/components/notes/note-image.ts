const MAX_NOTE_IMAGE_SIZE = 10 << 20

interface NoteImageContext {
  root: string
  notePath: string
}

interface NoteImageControllerOptions {
  getContext: () => NoteImageContext
  saveNoteImage: (root: string, notePath: string, data: string) => Promise<string>
  readNoteImage: (root: string, notePath: string, path: string) => Promise<string>
  reportError: (message: string) => void
}

function toErrorMessage(error: unknown): string {
  return error instanceof Error ? error.message : String(error)
}

async function fileToBase64(file: File): Promise<string> {
  if (file.size > MAX_NOTE_IMAGE_SIZE)
    throw new Error('图片大小不能超过 10 MiB')
  const bytes = new Uint8Array(await file.arrayBuffer())
  const chunks: string[] = []
  const chunkSize = 0x8000
  for (let offset = 0; offset < bytes.length; offset += chunkSize) {
    const chunk = bytes.subarray(offset, offset + chunkSize)
    let binary = ''
    for (let index = 0; index < chunk.length; index++)
      binary += String.fromCharCode(chunk[index]!)
    chunks.push(binary)
  }
  return btoa(chunks.join(''))
}

function isLocalImageSource(source: string): boolean {
  if (!source.trim() || source.startsWith('/') || source.startsWith('\\\\'))
    return false
  return !/^(?:[a-z][a-z\d+.-]*:|\/\/)/i.test(source)
}

// createNoteImageController creates note image controller.
export function createNoteImageController(options: NoteImageControllerOptions) {
  const pendingUploads = new Set<Promise<unknown>>()
  const previewCache = new Map<string, Promise<string>>()
  const reportedLoadErrors = new Set<string>()
  const reportedPreviewErrors = new Set<string>()

  function onUpload(file: File): Promise<string> {
    const { root, notePath } = options.getContext()
    const upload = (async () => {
      try {
        if (!root || !notePath)
          throw new Error('请先选择工作区和笔记后再上传图片')
        return await options.saveNoteImage(root, notePath, await fileToBase64(file))
      }
      catch (error) {
        options.reportError(`图片上传失败：${toErrorMessage(error)}`)
        throw error
      }
    })()
    return trackUpload(upload)
  }

  function trackUpload<T>(upload: Promise<T>): Promise<T> {
    const tracked = upload.finally(() => pendingUploads.delete(tracked))
    pendingUploads.add(tracked)
    return tracked
  }

  function proxyDomURL(source: string): Promise<string> | string {
    if (!isLocalImageSource(source))
      return source

    const { root, notePath } = options.getContext()
    const key = JSON.stringify([root, notePath, source])
    let preview = previewCache.get(key)
    if (!preview) {
      preview = (async () => {
        if (!root || !notePath)
          throw new Error('请先选择工作区和笔记后再加载图片')
        return await options.readNoteImage(root, notePath, source)
      })().catch((error) => {
        previewCache.delete(key)
        if (!reportedPreviewErrors.has(key)) {
          reportedPreviewErrors.add(key)
          options.reportError(`图片加载失败：${toErrorMessage(error)}`)
        }
        throw error
      })
      previewCache.set(key, preview)
    }
    return preview
  }

  function onImageLoadError(event: Event): void {
    const target = event.target as (EventTarget & { getAttribute?: (name: string) => string | null }) | null
    const source = target?.getAttribute?.('src') ?? ''
    if (!source || isLocalImageSource(source))
      return
    if (reportedLoadErrors.has(source))
      return
    reportedLoadErrors.add(source)
    options.reportError('图片加载失败，已保留原始图片链接')
  }

  async function waitForUploads(): Promise<void> {
    while (pendingUploads.size > 0)
      await Promise.allSettled([...pendingUploads])
  }

  function hasPendingUploads(): boolean {
    return pendingUploads.size > 0
  }

  function clearCache(): void {
    previewCache.clear()
    reportedLoadErrors.clear()
    reportedPreviewErrors.clear()
  }

  return {
    config: {
      onUpload,
      proxyDomURL,
      onImageLoadError,
      inlineUploadButton: '上传图片',
      inlineConfirmButton: '确认',
      inlineUploadPlaceholderText: '或粘贴图片链接',
      blockUploadButton: '上传图片',
      blockConfirmButton: '确认',
      blockUploadPlaceholderText: '或粘贴图片链接',
      blockCaptionPlaceholderText: '输入图片描述',
    },
    waitForUploads,
    hasPendingUploads,
    clearCache,
  }
}
