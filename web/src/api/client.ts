import createClient from 'openapi-fetch';
import type { paths } from './schema';
import { clearSession } from '#/features/auth/lib/session';

const BASE_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:8080';

const client = createClient<paths>({
  baseUrl: BASE_URL,
});

// Save a clone of each request before the body is consumed so we can retry
// after a token refresh.
const requestClones = new WeakMap<Request, Request>();

client.use({
  onRequest({ request }) {
    requestClones.set(request, request.clone());

    const token = localStorage.getItem('token');
    if (token) {
      request.headers.set('Authorization', `Bearer ${token}`);
    }
    return request;
  },
  async onResponse({ response, request }) {
    if (response.status !== 401 || request.url.includes('/auth/')) {
      return response;
    }

    const refreshToken = localStorage.getItem('refreshToken');
    if (!refreshToken) {
      clearSession();
      return response;
    }

    const refreshResponse = await fetch(`${BASE_URL}/auth/refresh`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${refreshToken}` },
    });

    if (!refreshResponse.ok) {
      clearSession();
      return response;
    }

    const data = await refreshResponse.json() as { token: string; refreshToken: string };
    localStorage.setItem('token', data.token);
    localStorage.setItem('refreshToken', data.refreshToken);

    const saved = requestClones.get(request);
    if (!saved) {
      clearSession();
      return response;
    }

    const headers = new Headers(saved.headers);
    headers.set('Authorization', `Bearer ${data.token}`);
    return fetch(new Request(saved, { headers }));
  },
});

export default client;
