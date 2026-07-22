export const fetchAndDownload = async (
  path: string,
  fallbackFilename: string,
): Promise<void> => {
  const response = await fetch(`/api${path}`, {
    credentials: 'include',
  });

  if (!response.ok) {
    throw new Error(`Export failed: ${response.statusText}`);
  }

  const disposition = response.headers.get('Content-Disposition') ?? '';
  const match = disposition.match(/filename=([^\s;]+)/);
  const filename = match ? match[1] : fallbackFilename;
  const blob = await response.blob();
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = filename;
  a.click();
  URL.revokeObjectURL(url);
};
