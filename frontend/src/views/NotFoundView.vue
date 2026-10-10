<script setup lang="ts">
import { RouterLink } from 'vue-router'
</script>

<template>
  <div class="py-20 text-center">
    <p class="text-5xl font-bold text-slate-300">404</p>
    <h1 class="mt-4 text-xl font-semibold text-slate-800">Page not found</h1>
    <p class="mt-2 text-sm text-slate-500">The page you are looking for does not exist.</p>
    <RouterLink
      :to="{ name: 'home' }"
      class="mt-6 inline-block rounded-lg bg-indigo-600 px-4 py-2 text-sm font-semibold text-white hover:bg-indigo-700"
    >
      Back to the feed
    </RouterLink>
  </div>
</template>
