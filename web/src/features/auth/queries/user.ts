import client from '#/api/client';
import type { components } from '#/api/schema';
import { queryOptions } from '@tanstack/react-query';

export type VisibleUser = components['schemas']['VisibleUser'];

export const userKeys = {
  me: ['user', 'me'] as const,
};

export const userOptions = queryOptions({
  queryKey: userKeys.me,
  queryFn: async () => {
    const { data, error } = await client.GET('/user');
    if (error) {
      throw error;
    }
    return data;
  },
});