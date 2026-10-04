<template>
  <n-layout has-sider class="h-screen">
    <AppSidebar />
    <n-layout-content content-style="padding: 16px 16px 16px 0;">
      <div class="rounded-xl bg-surface wh-full overflow-hidden">
        <n-scrollbar>
          <slot />
        </n-scrollbar>
      </div>
    </n-layout-content>
  </n-layout>
</template>
