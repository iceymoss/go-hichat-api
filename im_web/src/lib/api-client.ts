// Backend API client for Go services

const BACKEND_BASE = process.env.BACKEND_API_URL || 'http://127.0.0.1:8887';
const SOCIAL_BASE = process.env.SOCIAL_API_URL || 'http://127.0.0.1:8888';
const TREND_BASE = process.env.TREND_API_URL || 'http://127.0.0.1:8891';

export interface BackendResp<T = unknown> {
  code: number;
  msg: string;
  data: T;
}

export interface BackendUser {
  id: string;
  mobile: string;
  nickname: string;
  sex: number;
  avatar: string;
  lastLogin: string;
  introduction: string;
  email: string;
  region: string;
  occupation: string;
  tags: string;
}

export async function backendFetch<T = unknown>(
  path: string,
  options: RequestInit = {},
): Promise<BackendResp<T>> {
  const url = `${BACKEND_BASE}${path}`;
  const res = await fetch(url, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options.headers,
    },
  });
  return res.json() as Promise<BackendResp<T>>;
}

export async function backendGet<T = unknown>(
  path: string,
  token?: string,
): Promise<BackendResp<T>> {
  return backendFetch<T>(path, {
    method: 'GET',
    headers: token ? { Authorization: `Bearer ${token}` } : {},
  });
}

export async function backendPost<T = unknown>(
  path: string,
  body: unknown,
  token?: string,
): Promise<BackendResp<T>> {
  return backendFetch<T>(path, {
    method: 'POST',
    body: JSON.stringify(body),
    headers: token ? { Authorization: `Bearer ${token}` } : {},
  });
}

export async function backendPut<T = unknown>(
  path: string,
  body: unknown,
  token?: string,
): Promise<BackendResp<T>> {
  return backendFetch<T>(path, {
    method: 'PUT',
    body: JSON.stringify(body),
    headers: token ? { Authorization: `Bearer ${token}` } : {},
  });
}

// Social service clients (port 8888)
export async function socialGet<T = unknown>(
  path: string,
  token?: string,
): Promise<BackendResp<T>> {
  const url = `${SOCIAL_BASE}${path}`;
  const res = await fetch(url, {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
  });
  return res.json() as Promise<BackendResp<T>>;
}

export async function socialPost<T = unknown>(
  path: string,
  body: unknown,
  token?: string,
): Promise<BackendResp<T>> {
  const url = `${SOCIAL_BASE}${path}`;
  const res = await fetch(url, {
    method: 'POST',
    body: JSON.stringify(body),
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
  });
  return res.json() as Promise<BackendResp<T>>;
}

export async function socialPut<T = unknown>(
  path: string,
  body: unknown,
  token?: string,
): Promise<BackendResp<T>> {
  const url = `${SOCIAL_BASE}${path}`;
  const res = await fetch(url, {
    method: 'PUT',
    body: JSON.stringify(body),
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
  });
  return res.json() as Promise<BackendResp<T>>;
}

// Trend service client (port 8891)
export async function trendGet<T = unknown>(
  path: string,
  token?: string,
): Promise<BackendResp<T>> {
  const url = `${TREND_BASE}${path}`;
  const res = await fetch(url, {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
  });
  return res.json() as Promise<BackendResp<T>>;
}
