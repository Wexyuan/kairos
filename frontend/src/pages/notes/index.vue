<script setup lang="ts">
const treeCollapsed = ref(false)
</script>

<template>
  <div class="p-6 flex h-full transition-all duration-200" :class="treeCollapsed ? 'gap-0' : 'gap-6'">
    <div class="shrink-0 transition-all duration-200 overflow-hidden" :class="treeCollapsed ? 'w-0' : 'w-70'">
      <NoteTree class="h-full w-70" />
    </div>
    <NoteEditor class="flex-1 min-w-0" :tree-collapsed="treeCollapsed" @toggle-tree="treeCollapsed = !treeCollapsed" />
  </div>
</template>
