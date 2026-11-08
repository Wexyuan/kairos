import type { NoteNode } from '@/services/bindings'
import { NotesService } from '@/services/bindings'
import { STORAGE_KEYS } from '@/stores/app'

// toNoteNodes converts nullable nodes into a non-null note node array.
export function toNoteNodes(nodes: (NoteNode | null)[] | null): NoteNode[] {
  return (nodes ?? []).filter((node): node is NoteNode => node !== null)
}

// isInside reports whether a path is the target path or a descendant of it.
export function isInside(path: string, target: string): boolean {
  return path === target || path.startsWith(`${target}/`)
}

// useNotesStore defines a store for managing notes in the workspace.
export const useNotesStore = defineStore('notes', () => {
  const root = useLocalStorage(STORAGE_KEYS.notesWorkspace, '')
  const tree = ref<NoteNode[]>([])
  const activePath = ref('')
  const treeLoading = ref(false)

  let editorFlush: (() => Promise<boolean>) | null = null

  // workspaceName returns the name of the workspace path.
  const workspaceName = computed(() => {
    const parts = root.value.split(/[\\/]/).filter(Boolean)
    return parts.at(-1) ?? ''
  })

  // nodeIndex returns a map of note nodes by their workspace relative path for quick lookup.
  const nodeIndex = computed(() => {
    const index = new Map<string, NoteNode>()
    const walk = (nodes: NoteNode[]) => {
      for (const node of nodes) {
        index.set(node.path, node)
        walk(toNoteNodes(node.children))
      }
    }
    walk(tree.value)
    return index
  })

  // registerEditorFlush registers a editor flush hook.
  function registerEditorFlush(flush: (() => Promise<boolean>) | null): void {
    editorFlush = flush
  }

  // findNode returns the note node for a given path or null if not found.
  function findNode(path: string): NoteNode | null {
    return nodeIndex.value.get(path) ?? null
  }

  // selectWorkspace selects a system folder as the workspace.
  async function selectWorkspace(): Promise<boolean> {
    const picked = await NotesService.SelectWorkspace()
    if (!picked)
      return false
    await flushEditor()
    root.value = picked
    activePath.value = ''
    await listTree()
    return true
  }

  // listTree lists the note file tree under a workspace folder.
  async function listTree(): Promise<void> {
    if (!root.value)
      return
    treeLoading.value = true
    try {
      tree.value = toNoteNodes(await NotesService.ListTree(root.value))
    }
    finally {
      treeLoading.value = false
    }
  }

  // createFolder creates a new note folder.
  async function createFolder(path: string): Promise<void> {
    await NotesService.CreateFolder({ root: root.value, path })
    await listTree()
  }

  // createNote creates a new markdown note.
  async function createNote(path: string): Promise<void> {
    await NotesService.CreateNote({ root: root.value, path })
    await listTree()
    activePath.value = path
  }

  // saveNote saves the markdown content of a note.
  async function saveNote(path: string, content: string, targetRoot = root.value): Promise<void> {
    await NotesService.SaveNote({ root: targetRoot, path, content })
  }

  // saveNoteImage saves base64 image data and returns its note-relative markdown path.
  async function saveNoteImage(root: string, notePath: string, data: string): Promise<string> {
    const response = await NotesService.SaveNoteImage({ root, note_path: notePath, data })
    if (!response?.path)
      throw new Error('后端未返回图片路径')
    return response.path
  }

  // readNote reads the markdown content of a note.
  async function readNote(path: string, targetRoot = root.value): Promise<string> {
    return await NotesService.ReadNote({ root: targetRoot, path })
  }

  // readNoteImage reads an image and returns a data URL for preview only.
  async function readNoteImage(root: string, notePath: string, path: string): Promise<string> {
    const response = await NotesService.ReadNoteImage({ root, note_path: notePath, path })
    if (!response)
      throw new Error('后端未返回图片数据')
    return `data:${response.mime};base64,${response.data}`
  }

  // renameNote renames a note or directory.
  async function renameNote(oldPath: string, newPath: string): Promise<void> {
    await flushEditor()
    await NotesService.RenameNote({ root: root.value, old_path: oldPath, new_path: newPath })
    await listTree()
    if (isInside(activePath.value, oldPath))
      activePath.value = newPath + activePath.value.slice(oldPath.length)
  }

  // deleteNote deletes a note or directory.
  async function deleteNote(path: string): Promise<void> {
    await flushEditor()
    await NotesService.DeleteNote({ root: root.value, path })
    await listTree()
    if (isInside(activePath.value, path))
      activePath.value = ''
  }

  // flushEditor flushes the editor.
  async function flushEditor(): Promise<void> {
    if (!editorFlush)
      return
    const saved = await editorFlush()
    if (!saved)
      throw new Error('存在未保存的修改')
  }

  return {
    root,
    tree,
    activePath,
    treeLoading,
    workspaceName,
    registerEditorFlush,
    findNode,
    selectWorkspace,
    listTree,
    createFolder,
    createNote,
    saveNote,
    saveNoteImage,
    readNote,
    readNoteImage,
    renameNote,
    deleteNote,
  }
})
