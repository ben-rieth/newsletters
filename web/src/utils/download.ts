import { refreshSession } from '#/api/client';
import { clearSession } from '#/features/auth/lib/session';

const request = (path: string) =>
  fetch(`/api${path}`, { credentials: 'include' });

export const fetchAndDownload = async (
  path: string,
  fallbackFilename: string,
): Promise<void> => {
  let response = await request(path);

  if (response.status === 401) {
    if (!(await refreshSession())) {
      clearSession();
      throw new Error(`Export failed (${response.status})`);
    }
    response = await request(path);
  }

  if (!response.ok) {
    throw new Error(`Export failed (${response.status})`);
  }

  const disposition = response.headers.get('Content-Disposition') ?? '';
  const match = disposition.match(/filename="?([^";]+)"?/);
  const filename = match ? match[1] : fallbackFilename;
  const blob = await response.blob();
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = filename;
  a.click();
  URL.revokeObjectURL(url);
};
