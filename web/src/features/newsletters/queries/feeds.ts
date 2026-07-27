import client from '#/api/client';
import type { components } from '#/api/schema';
import { queryOptions } from '@tanstack/react-query';

export type Feed = components['schemas']['UiFeed'];
export type FeedDetail = components['schemas']['UiDetailedFeed'];
export type FeedMetaData = components['schemas']['FeedMetaData'];
export type ItemPreview = components['schemas']['ItemPreview'];
export type FeedFilter = components['schemas']['FeedFilter'];
export type FeedHealth = components['schemas']['FeedHealth'];

export const feedsKeys = {
  list: (newsletterId: string) => ['feeds', newsletterId] as const,
  metadata: (url: string) => ['feedMetadata', url] as const,
};

export const feedDetailKeys = {
  detail: (newsletterId: string, feedId: string) =>
    ['feedDetail', newsletterId, feedId] as const,
};

export const feedPreviewKeys = {
  preview: (newsletterId: string, feedId: string) =>
    ['feedPreview', newsletterId, feedId] as const,
};

export const feedsOptions = (newsletterId: string) =>
  queryOptions({
    queryKey: feedsKeys.list(newsletterId),
    queryFn: async () => {
      const { data, error } = await client.GET(
        '/newsletter/{newsletterId}/feed',
        { params: { path: { newsletterId } } },
      );
      if (error) {
        throw error;
      }
      return data ?? [];
    },
  });

export const feedDetailOptions = (newsletterId: string, feedId: string) =>
  queryOptions({
    queryKey: feedDetailKeys.detail(newsletterId, feedId),
    queryFn: async () => {
      const { data, error } = await client.GET(
        '/newsletter/{newsletterId}/feed/{feedId}',
        { params: { path: { newsletterId, feedId } } },
      );
      if (error) {
        throw error;
      }
      return data;
    },
  });

export const feedPreviewOptions = (newsletterId: string, feedId: string) =>
  queryOptions({
    queryKey: feedPreviewKeys.preview(newsletterId, feedId),
    queryFn: async () => {
      const { data, error } = await client.GET(
        '/newsletter/{newsletterId}/feed/{feedId}/preview',
        { params: { path: { newsletterId, feedId } } },
      );
      if (error) {
        throw error;
      }
      return data ?? [];
    },
    enabled: false,
  });

export const feedMetadataOptions = (url: string) =>
  queryOptions({
    queryKey: feedsKeys.metadata(url),
    queryFn: async () => {
      const { data, error } = await client.POST('/feed', { body: { url } });
      if (error) {
        throw error;
      }
      return data;
    },
    enabled: !!url,
    retry: false,
    staleTime: 5 * 60 * 1000,
  });
