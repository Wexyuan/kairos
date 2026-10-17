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
const treeOptions = computed<TreeOption[]>(() => toTreeOptions(store.tree))
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
  return h('span', { class: [cls, 'h-4 w-4', selected ? 'text-brand' : 'text-ink'] })
}

function overrideNodeClick({ option }: { option: TreeOption }): TreeOverrideNodeClickBehaviorReturn {
  return option.isLeaf ? 'default' : 'toggleExpand'
}

function nodeProps({ option }: { option: TreeOption }) {
  return {
    'data-path': String(option.key),
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
  <div class="pa-2 rounded-xl bg-page flex flex-col gap-2 h-full overflow-hidden">
    <div class="flex shrink-0 gap-2 items-center">
      <n-button quaternary size="small" class="icon-btn shrink-0" :class="{ '!bg-brand/10 !text-brand': store.root }">
        <template #icon>
          <span class="i-mdi-folder-outline h-4 w-4" />
        </template>
      </n-button>
      <div class="title-sm flex-1 min-w-0 truncate" :title="store.root || '选择工作区开始记录叭 (^v^)'">
        {{ store.workspaceName || '未选择工作区' }}
      </div>
      <n-spin v-if="store.loading" :size="14" class="shrink-0" />
      <n-button quaternary size="small" class="icon-btn" title="选择工作区" @click="selectWorkspace">
        <template #icon>
          <span class="i-mdi-folder-search-outline h-4 w-4" />
        </template>
      </n-button>
      <n-button quaternary size="small" class="icon-btn" title="定位当前笔记" :disabled="!store.activePath" @click="locateActiveNote">
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
      <n-spin v-if="store.loading" :size="24" />
      <n-empty v-else description="暂无笔记，快去新建叭 (^v^)">
        <template #extra>
          <n-button size="small" type="primary" @click="openCreateNoteDialog('')">
            新建笔记
          </n-button>
        </template>
      </n-empty>
    </div>

    <div v-else ref="treeContainerRef" class="flex flex-1 flex-col min-h-0">
      <n-scrollbar class="flex-1">
        <n-tree
          v-model:expanded-keys="expandedKeys"
          block-line
          ellipsis
          :data="treeOptions"
          :render-prefix="renderPrefix"
          :override-default-node-click-behavior="overrideNodeClick"
          :node-props="nodeProps"
          :selected-keys="selectedKeys"
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
