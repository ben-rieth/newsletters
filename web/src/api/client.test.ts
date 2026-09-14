import { describe, expect, it, vi, beforeEach, afterEach } from 'vitest';
import { refreshSession } from './client';

describe('refreshSession', () => {
  beforeEach(() => {
    vi.stubGlobal('fetch', vi.fn());
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  const respondWith = (ok: boolean) => {
    let release!: (value: Response) => void;
    const pending = new Promise<Response>((resolve) => {
      release = resolve;
    });

    vi.mocked(fetch).mockReturnValue(pending);

    return () => release({ ok } as Response);
  };

  it('makes a single request for concurrent callers', async () => {
    const release = respondWith(true);

    const calls = [refreshSession(), refreshSession(), refreshSession()];
    release();

    expect(await Promise.all(calls)).toEqual([true, true, true]);
    expect(fetch).toHaveBeenCalledTimes(1);
  });

  it('starts a new request once the previous one settles', async () => {
    const releaseFirst = respondWith(true);
    const first = refreshSession();
    releaseFirst();
    await first;

    const releaseSecond = respondWith(true);
    const second = refreshSession();
    releaseSecond();
    await second;

    expect(fetch).toHaveBeenCalledTimes(2);
  });

  it('reports failure to every concurrent caller', async () => {
    const release = respondWith(false);

    const calls = [refreshSession(), refreshSession()];
    release();

    expect(await Promise.all(calls)).toEqual([false, false]);
    expect(fetch).toHaveBeenCalledTimes(1);
  });

  it('treats a network error as a failed refresh rather than throwing', async () => {
    vi.mocked(fetch).mockRejectedValue(new Error('offline'));

    await expect(refreshSession()).resolves.toBe(false);
  });
});
