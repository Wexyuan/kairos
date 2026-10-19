<template>
  <n-layout has-sider class="h-screen">
    <AppSidebar />
    <n-layout-content>
      <div class="bg-surface wh-full overflow-hidden">
        <n-scrollbar content-style="height: 100%">
          <slot />
        </n-scrollbar>
      </div>
    </n-layout-content>
  </n-layout>
</template>
