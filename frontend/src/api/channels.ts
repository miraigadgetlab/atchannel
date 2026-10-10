import { request } from './client'
import type { Channel, ChannelDetail, PageParams, Paginated } from '@/types/api'

export interface ChannelListParams extends PageParams {
  q?: string
}

export interface CreateChannelInput {
  name: string
  title: string
  description: string
}

export interface UpdateChannelInput {
  name?: string
  title?: string
  description?: string
}

export const channelsApi = {
  list(params: ChannelListParams = {}): Promise<Paginated<Channel>> {
    return request('/channels', {
      query: { q: params.q, limit: params.limit, offset: params.offset },
    })
  },

  get(id: number | string, params: PageParams = {}): Promise<ChannelDetail> {
    return request(`/channels/${id}`, {
      query: { limit: params.limit, offset: params.offset },
    })
  },

  create(input: CreateChannelInput): Promise<Channel> {
    return request('/channels', { method: 'POST', body: input })
  },

  update(id: number | string, input: UpdateChannelInput): Promise<Channel> {
    return request(`/channels/${id}`, { method: 'PUT', body: input })
  },

  remove(id: number | string): Promise<{ deleted: boolean }> {
    return request(`/channels/${id}`, { method: 'DELETE' })
  },
}
