export const NEWSLETTER_TABS = ['feeds', 'history', 'settings'] as const;
export type NewsletterTab = (typeof NEWSLETTER_TABS)[number];

export const FEED_TABS = ['details', 'filters'] as const;
export type FeedTab = (typeof FEED_TABS)[number];

export const parseTab = <T extends string>(
  tabs: readonly T[],
  value: unknown,
): T | undefined => tabs.find((tab) => tab === value);
