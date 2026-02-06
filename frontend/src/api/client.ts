import axios from 'axios';
import { authStore } from '../store/auth';

const baseURL = import.meta.env.VITE_API_URL || '';

const api = axios.create({
  baseURL
});

const refreshClient = axios.create({
  baseURL
});

api.interceptors.request.use((config) => {
  const token = authStore.getAccessToken();
  if (token) {
    config.headers = config.headers ?? {};
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

let refreshPromise: Promise<string> | null = null;

async function refreshTokens(): Promise<string> {
  const refreshToken = authStore.getRefreshToken();
  if (!refreshToken) {
    throw new Error('Missing refresh token');
  }
  const { data } = await refreshClient.post<{ access_token: string; refresh_token: string }>('/auth/refresh', {
    refresh_token: refreshToken
  });
  authStore.setTokens(data.access_token, data.refresh_token);
  return data.access_token;
}

api.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config as (typeof error.config & { _retry?: boolean }) | undefined;
    if (!originalRequest || originalRequest._retry) {
      return Promise.reject(error);
    }

    if (error.response?.status === 401 && !String(originalRequest.url || '').includes('/auth/refresh')) {
      originalRequest._retry = true;
      try {
        if (!refreshPromise) {
          refreshPromise = refreshTokens().finally(() => {
            refreshPromise = null;
          });
        }
        const newAccessToken = await refreshPromise;
        originalRequest.headers = originalRequest.headers ?? {};
        originalRequest.headers.Authorization = `Bearer ${newAccessToken}`;
        return api(originalRequest);
      } catch (refreshError) {
        authStore.clear();
        return Promise.reject(refreshError);
      }
    }

    return Promise.reject(error);
  }
);

export default api;
