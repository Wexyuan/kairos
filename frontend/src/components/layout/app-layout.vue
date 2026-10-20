<template>
  <n-layout has-sider class="h-screen">
    <AppSidebar />
    <n-layout-content>
      <div class="flex flex-col h-full">
        <AppTopbar />
        <div class="bg-surface flex-1 min-h-0 overflow-hidden">
          <n-scrollbar content-style="height: 100%">
            <slot />
          </n-scrollbar>
        </div>
      </div>
    </n-layout-content>
  </n-layout>
</template>
