<script setup lang="ts">
import { RouterView } from 'vue-router'

import AppNav from '@/components/AppNav.vue'
</script>

<template>
  <div class="flex min-h-screen flex-col">
    <AppNav />

    <main class="mx-auto w-full max-w-5xl flex-1 px-4 py-6">
      <RouterView />
    </main>

    <footer class="border-t border-slate-200 py-6 text-center text-xs text-slate-400">
      atchannel — a small discussion forum.
    </footer>
  </div>
</template>
