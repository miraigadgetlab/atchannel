<script setup lang="ts">
import { RouterView } from 'vue-router'

import AppNav from '@/components/AppNav.vue'
import { DISCORD_URL, GITHUB_URL } from '@/config'
</script>

<template>
  <div class="flex min-h-screen flex-col">
    <AppNav />

    <main class="mx-auto w-full max-w-5xl flex-1 px-4 py-6">
      <RouterView />
    </main>

    <footer class="border-t border-slate-200 py-6 text-xs text-slate-400">
      <!-- Deliberately understated: small, muted, no underline. Order is
           discord then github, matching the requested "discord github". -->
      <div class="mx-auto flex max-w-5xl items-center justify-end gap-4 px-4">
        <a
          v-if="DISCORD_URL"
          :href="DISCORD_URL"
          target="_blank"
          rel="noopener noreferrer"
          class="hover:text-slate-600"
        >
          discord
        </a>
        <a
          :href="GITHUB_URL"
          target="_blank"
          rel="noopener noreferrer"
          class="hover:text-slate-600"
        >
          github
        </a>
      </div>
    </footer>
  </div>
</template>
