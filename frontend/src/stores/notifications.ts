import { ref } from 'vue'
import { defineStore } from 'pinia'

import { notificationsApi } from '@/api/notifications'

/** Keeps the navbar badge in sync without every view fetching the count. */
export const useNotificationsStore = defineStore('notifications', () => {
  const unread = ref(0)

  async function refreshUnread(): Promise<void> {
    try {
      const result = await notificationsApi.unreadCount()
      unread.value = result.unread
    } catch {
      // Anonymous users get a 401 here; treat it as "no badge".
      unread.value = 0
    }
  }

  async function markAllRead(): Promise<void> {
    await notificationsApi.markAllRead()
    unread.value = 0
  }

  function reset(): void {
    unread.value = 0
  }

  return { unread, refreshUnread, markAllRead, reset }
})
