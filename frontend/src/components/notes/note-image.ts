const MAX_PREVIEW_CACHE_SIZE = 20
const MAX_PREVIEW_CACHE_BYTES = 32 << 20

interface NoteImageContext {
  root: string
  notePath: string
}

interface NoteImageControllerOptions {
  getContext: () => NoteImageContext
  getNoteMaxImageSize: () => number
  saveNoteImage: (root: string, notePath: string, data: string) => Promise<string>
  readNoteImage: (root: string, notePath: string, path: string) => Promise<string>
  reportError: (message: string) => void
}

function toErrorMessage(error: unknown): string {
  return error instanceof Error ? error.message : String(error)
}

async function fileToBase64(file: File, maxSize: number): Promise<string> {
  if (file.size > maxSize)
    throw new Error(`图片大小不能超过 ${formatSize(maxSize)}`)
  const bytes = new Uint8Array(await file.arrayBuffer())
  let binary = ''
  const chunkSize = 0x8000
  for (let offset = 0; offset < bytes.length; offset += chunkSize)
    binary += String.fromCharCode(...bytes.subarray(offset, offset + chunkSize))
  return btoa(binary)
}

function formatSize(bytes: number): string {
  const mib = bytes / (1 << 20)
  return `${Number.isInteger(mib) ? mib : mib.toFixed(1)} MiB`
}

function isLocalImageSource(source: string): boolean {
  if (!source.trim() || source.startsWith('/') || source.startsWith('\\'))
    return false
  return !/^(?:[a-z][a-z\d+.-]*:|\/\/)/i.test(source)
}

// createNoteImageController creates note image controller.
export function createNoteImageController(options: NoteImageControllerOptions) {
  const pendingUploads = new Set<Promise<unknown>>()
  const previewCache = new Map<string, Promise<string>>()
  const previewSizes = new Map<string, number>()
  const reportedLoadErrors = new Set<string>()
  const reportedPreviewErrors = new Set<string>()
  let previewBytes = 0

  function onUpload(file: File): Promise<string> {
    const { root, notePath } = options.getContext()
    const upload = (async () => {
      try {
        if (!root || !notePath)
          throw new Error('请先选择工作区和笔记后再上传图片')
        return await options.saveNoteImage(root, notePath, await fileToBase64(file, options.getNoteMaxImageSize()))
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
    const cached = previewCache.get(key)
    if (cached) {
      // refresh recency for LRU behavior.
      previewCache.delete(key)
      previewCache.set(key, cached)
      return cached
    }

    const preview = (async () => {
      if (!root || !notePath)
        throw new Error('请先选择工作区和笔记后再加载图片')
      return await options.readNoteImage(root, notePath, source)
    })().catch((error) => {
      dropPreview(key)
      if (!reportedPreviewErrors.has(key)) {
        reportedPreviewErrors.add(key)
        options.reportError(`图片加载失败：${toErrorMessage(error)}`)
      }
      throw error
    })
    previewCache.set(key, preview)
    preview.then((url) => {
      if (previewCache.get(key) !== preview)
        return
      previewSizes.set(key, url.length)
      previewBytes += url.length
      trimPreviewCache()
    }, () => {})
    trimPreviewCache()
    return preview
  }

  function dropPreview(key: string): void {
    previewCache.delete(key)
    const size = previewSizes.get(key)
    if (size === undefined)
      return
    previewBytes -= size
    previewSizes.delete(key)
  }

  function trimPreviewCache(): void {
    while (previewCache.size > MAX_PREVIEW_CACHE_SIZE || previewBytes > MAX_PREVIEW_CACHE_BYTES) {
      const oldest = previewCache.keys().next().value
      if (oldest === undefined)
        break
      dropPreview(oldest)
    }
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
    previewSizes.clear()
    previewBytes = 0
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
