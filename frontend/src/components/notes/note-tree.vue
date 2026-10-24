<script setup lang="ts">
import type { DropdownOption, TreeOption, TreeOverrideNodeClickBehaviorReturn } from 'naive-ui'
import type { VNodeChild } from 'vue'
import type { NoteNode } from '@/services/bindings'
import { basename, dirname, join } from 'pathe'
import { useAction } from '@/composables'
import { toNoteNodes, useNotesStore } from '@/stores/notes'

const store = useNotesStore()
const { runAction, confirmAction, inputAction } = useAction()

const treeContainerRef = ref<HTMLElement | null>(null)
const searchText = ref('')
const filteredTree = computed(() => filterNotes(store.tree, searchText.value.trim().toLowerCase()))
const treeOptions = computed<TreeOption[]>(() => toTreeOptions(filteredTree.value))
const selectedKeys = computed(() => (store.activePath ? [store.activePath] : []))
const expandedKeys = ref<Array<string | number>>([])
const contextMenu = ref({ show: false, x: 0, y: 0, path: '' })
const noteNodeActions: DropdownOption[] = [
  { key: 'create', label: '新建笔记', icon: renderIcon('i-mdi-file-plus-outline') },
  { key: 'rename', label: '重命名', icon: renderIcon('i-mdi-pencil-outline') },
  { key: 'delete', label: '删除', icon: renderIcon('i-mdi-delete-outline') },
]

onMounted(() => runAction('加载笔记', () => store.listTree()))

watch(() => store.activePath, (path) => {
  if (path)
    expandAncestors(path)
})

watch(searchText, (value) => {
  if (!value.trim())
    return
  expandedKeys.value = [...new Set([...expandedKeys.value, ...collectDirKeys(filteredTree.value)])]
})

async function selectWorkspace(): Promise<void> {
  let picked = false
  const loaded = await runAction('加载笔记', async () => {
    picked = await store.selectWorkspace()
  })
  if (picked && loaded)
    expandedKeys.value = []
}

function locateActiveNote(): void {
  const path = store.activePath
  if (!path)
    return
  expandAncestors(path)
  nextTick(() => {
    treeContainerRef.value?.querySelector(`[data-path="${CSS.escape(path)}"]`)?.scrollIntoView({ block: 'center', behavior: 'smooth' })
  })
}

function toTreeOptions(nodes: NoteNode[]): TreeOption[] {
  return nodes.map((node) => {
    const option: TreeOption = { key: node.path, label: basename(node.path), isLeaf: !node.is_dir }
    if (node.is_dir)
      option.children = toTreeOptions(toNoteNodes(node.children))
    return option
  })
}

function renderPrefix({ option, selected }: { option: TreeOption, selected: boolean }): VNodeChild {
  const expanded = expandedKeys.value.includes(String(option.key))
  const cls = option.isLeaf ? 'i-mdi-file-document-outline' : expanded ? 'i-mdi-folder-open-outline' : 'i-mdi-folder-outline'
  return h('span', { class: [cls, 'h-4 w-4', selected ? 'text-brand' : 'text-ink/60'] })
}

function overrideNodeClick({ option }: { option: TreeOption }): TreeOverrideNodeClickBehaviorReturn {
  return option.isLeaf ? 'default' : 'toggleExpand'
}

function nodeProps({ option }: { option: TreeOption }) {
  return {
    'data-path': String(option.key),
    'title': String(option.label),
    'onContextmenu': (e: MouseEvent) => {
      e.preventDefault()
      e.stopPropagation()
      openContextMenu(e, String(option.key))
    },
  }
}

function openContextMenu(event: MouseEvent, path: string): void {
  contextMenu.value = { show: true, x: event.clientX, y: event.clientY, path }
}

function filterNotes(nodes: NoteNode[], keyword: string): NoteNode[] {
  if (!keyword)
    return nodes
  const result: NoteNode[] = []
  for (const node of nodes) {
    const name = basename(node.path).toLowerCase()
    if (node.is_dir) {
      const children = filterNotes(toNoteNodes(node.children), keyword)
      if (name.includes(keyword) || children.length)
        result.push({ ...node, children })
    }
    else if (name.includes(keyword)) {
      result.push(node)
    }
  }
  return result
}

function collectDirKeys(nodes: NoteNode[]): string[] {
  const keys: string[] = []
  for (const node of nodes) {
    if (!node.is_dir)
      continue
    keys.push(node.path)
    keys.push(...collectDirKeys(toNoteNodes(node.children)))
  }
  return keys
}

function handleNoteNodeSelect(keys: Array<string | number>): void {
  store.activePath = keys.length > 0 ? String(keys[0]) : ''
}

function expandAncestors(path: string): void {
  const expanded = new Set(expandedKeys.value)
  let prefix = ''
  for (const segment of path.split('/').slice(0, -1)) {
    prefix = prefix ? `${prefix}/${segment}` : segment
    expanded.add(prefix)
  }
  expandedKeys.value = [...expanded]
}

function handleContextMenuSelect(key: string | number): void {
  const { path } = contextMenu.value
  contextMenu.value.show = false
  switch (key) {
    case 'create': {
      const node = store.findNode(path)
      openCreateNoteDialog(!node || node.is_dir ? path : dirname(path))
      break
    }
    case 'rename':
      openRenameNoteDialog(path)
      break
    case 'delete':
      openDeleteNoteDialog(path)
      break
  }
}

function openCreateNoteDialogInCurFolder(): void {
  openCreateNoteDialog(store.activePath ? dirname(store.activePath) : '')
}

function openCreateNoteDialog(dir: string): void {
  inputAction({ title: '新建笔记', initial: getDefaultNoteName(dir), action: name => submitCreateNote(dir, name) })
}

async function submitCreateNote(dir: string, name: string): Promise<void> {
  if (!name)
    throw new Error('请输入笔记名称')
  const path = join(dir, ensureNoteExt(name))
  if (store.findNode(path))
    throw new Error(`已存在同名笔记：${basename(path)}`)
  await store.createNote(path)
}

function openRenameNoteDialog(path: string): void {
  inputAction({ title: '重命名', initial: basename(path), action: name => submitRenameNote(path, name) })
}

async function submitRenameNote(path: string, name: string): Promise<void> {
  if (!name)
    throw new Error('请输入笔记名称')
  const target = join(dirname(path), ensureNoteExt(name))
  if (target === path)
    return
  if (store.findNode(target))
    throw new Error(`已存在同名笔记：${basename(target)}`)
  await store.renameNote(path, target)
}

function openDeleteNoteDialog(path: string): void {
  const isDir = store.findNode(path)?.is_dir ?? false
  confirmAction({
    title: '删除',
    content: isDir ? '是否删除文件夹及其全部内容？' : '是否删除笔记？',
    action: () => store.deleteNote(path),
  })
}

function renderIcon(icon: string): () => VNodeChild {
  return () => h('span', { class: `${icon} h-4 w-4` })
}

function ensureNoteExt(name: string): string {
  return /\.md$/i.test(name) ? name : `${name}.md`
}

function getDefaultNoteName(dir: string): string {
  let i = 0
  let name = '未命名.md'
  while (store.findNode(join(dir, name))) name = `未命名 ${++i}.md`
  return name
}
</script>

<template>
  <div class="panel flex flex-col h-full">
    <div class="px-3 py-2.5 border-b border-ink/5 flex gap-2 items-center">
      <n-button quaternary size="small" class="flex-1 min-w-0 justify-start" :title="store.root || '选择工作区'" @click="selectWorkspace">
        <template #icon>
          <span class="i-mdi-folder-outline h-4 w-4" :class="store.root ? 'text-brand' : 'text-ink/40'" />
        </template>
        <span class="title-sm min-w-0 truncate">{{ store.workspaceName || '未选择工作区' }}</span>
      </n-button>
      <n-spin v-if="store.treeLoading" :size="12" class="shrink-0" />
      <n-button type="primary" size="small" class="shrink-0" :disabled="!store.root" title="新建笔记" @click="openCreateNoteDialogInCurFolder">
        <template #icon>
          <span class="i-mdi-plus h-4 w-4" />
        </template>
        新建
      </n-button>
    </div>

    <div v-if="store.root" class="px-3 py-2.5 border-b border-ink/5 flex gap-2 items-center">
      <n-input v-model:value="searchText" size="small" clearable placeholder="搜索笔记…" class="flex-1 min-w-0">
        <template #prefix>
          <span class="i-mdi-magnify text-ink/40 h-4 w-4" />
        </template>
      </n-input>
      <n-button quaternary size="small" class="shrink-0" title="定位当前笔记" :disabled="!store.activePath" @click="locateActiveNote">
        <template #icon>
          <span class="i-mdi-crosshairs-gps h-4 w-4" />
        </template>
      </n-button>
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

    <div v-else-if="!store.tree.length" class="flex-center flex-1">
      <n-spin v-if="store.treeLoading" :size="24" />
      <n-empty v-else description="暂无笔记，先新建一篇叭 (^v^)">
        <template #extra>
          <n-button size="small" type="primary" @click="openCreateNoteDialog('')">
            新建笔记
          </n-button>
        </template>
      </n-empty>
    </div>

    <div v-else ref="treeContainerRef" class="flex-1 min-h-0">
      <n-scrollbar class="h-full">
        <div v-if="!treeOptions.length" class="title-sm text-ink/40 py-10 text-center">
          没有匹配的笔记
        </div>
        <n-tree
          v-else
          v-model:expanded-keys="expandedKeys"
          block-line
          ellipsis
          :data="treeOptions"
          :indent="14"
          :render-prefix="renderPrefix"
          :override-default-node-click-behavior="overrideNodeClick"
          :node-props="nodeProps"
          :selected-keys="selectedKeys"
          :cancelable="false"
          class="px-3 py-2.5"
          @update:selected-keys="handleNoteNodeSelect"
        />
      </n-scrollbar>
    </div>

    <n-dropdown
      trigger="manual"
      placement="bottom-start"
      :options="noteNodeActions"
      :show="contextMenu.show"
      :x="contextMenu.x"
      :y="contextMenu.y"
      @select="handleContextMenuSelect"
      @clickoutside="contextMenu.show = false"
    />
  </div>
</template>

<style scoped>
:deep(.n-input__prefix) {
  margin-right: 6px;
}

:deep(.n-tree-node--selected)::before {
  content: '';
  position: absolute;
  top: 50%;
  left: 0;
  width: 3px;
  height: 18px;
  border-radius: 999px;
  background: var(--app-brand);
  transform: translateY(-50%);
}

:deep(.n-tree-node--selected .n-tree-node-content__text) {
  color: var(--app-brand);
}
</style>
