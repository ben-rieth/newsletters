import client from '#/api/client'
import type { components } from '#/api/schema'
import { mutationOptions, queryOptions } from '@tanstack/react-query'

export type Newsletter = components['schemas']['Newsletter']

export type Frequency = 'daily' | 'weekly' | 'monthly'

export const newslettersKeys = {
  all: ['newsletters' as const],
  detail: (id: string) => ['newsletters', id] as const,
}

export const newslettersOptions = queryOptions({
  queryKey: newslettersKeys.all,
  queryFn: async () => {
    const { data, error } = await client.GET('/newsletters')
    if (error) {
      throw error
    }
    return data
  },
})

export const newsletterOptions = (newsletterId: string) => {
  return queryOptions({
    queryKey: newslettersKeys.detail(newsletterId),
    queryFn: async () => {
      const { data, error } = await client.GET('/newsletters/{newsletterId}', {
        params: { path: { newsletterId } },
      })
      if (error) {
        throw error
      }
      return data
    },
  })
}
