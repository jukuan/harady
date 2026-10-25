// Thin fetch wrappers for /api/admin/*. All requests include credentials so
// the session cookie is sent.

import { API_BASE } from '../config'

export class AdminError extends Error {
  status: number
  constructor(status: number, message: string) {
    super(message)
    this.status = status
  }
}

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    credentials: 'include',
    ...init,
    headers: {
      Accept: 'application/json',
      ...(init.body ? { 'Content-Type': 'application/json' } : {}),
      ...(init.headers ?? {}),
    },
  })
  if (res.status === 401) {
    throw new AdminError(401, 'unauthorized')
  }
  const text = await res.text()
  const data = text ? JSON.parse(text) : {}
  if (!res.ok) {
    throw new AdminError(res.status, data.error ?? 'request failed')
  }
  return data as T
}

// ---- auth -----------------------------------------------------------------

export async function requestMagicLink(email: string): Promise<void> {
  await request('/api/admin/auth/request', {
    method: 'POST',
    body: JSON.stringify({ email }),
  })
}

export async function fetchMe(): Promise<{ email: string }> {
  return request('/api/admin/auth/me')
}

export async function logout(): Promise<void> {
  await request('/api/admin/auth/logout', { method: 'POST' })
}

// ---- cities ---------------------------------------------------------------

export interface CityRow {
  id: number
  name: string
  region: string
  source?: string
  created_at: string
  updated_at: string
}

export async function listCities(q: string): Promise<CityRow[]> {
  const qs = q ? `?q=${encodeURIComponent(q)}` : ''
  const data = await request<{ cities: CityRow[] }>(`/api/admin/cities${qs}`)
  return data.cities
}

export async function createCity(name: string, region: string): Promise<void> {
  await request('/api/admin/cities', {
    method: 'POST',
    body: JSON.stringify({ name, region }),
  })
}

export async function updateCity(id: number, name: string, region: string): Promise<void> {
  await request(`/api/admin/cities/${id}`, {
    method: 'PUT',
    body: JSON.stringify({ name, region }),
  })
}

export async function deleteCity(id: number): Promise<void> {
  await request(`/api/admin/cities/${id}`, { method: 'DELETE' })
}

// ---- tables ---------------------------------------------------------------

export interface TableInfo {
  name: string
  rows: number
  columns: number
}

export interface TableDump {
  table: string
  columns: string[]
  rows: unknown[][]
  limit: number
  offset: number
  schema: string
}

export async function listTables(): Promise<TableInfo[]> {
  const data = await request<{ tables: TableInfo[] }>('/api/admin/tables')
  return data.tables
}

export async function dumpTable(name: string, limit = 100, offset = 0): Promise<TableDump> {
  return request(`/api/admin/tables/${encodeURIComponent(name)}?limit=${limit}&offset=${offset}`)
}

// ---- stats ----------------------------------------------------------------

export interface StatsSummary {
  cities: {
    total_cities: number
    total_submissions: number
    unique_cities_used: number
    never_used: number
    learned_cities: number
  }
  games: {
    total_games: number
    max_chain_length: number
    max_chain_human_only: number
    avg_chain_length: number
    total_moves: number
    games_with_bots: number
  }
}

export async function fetchSummary(): Promise<StatsSummary> {
  return request('/api/admin/stats/summary')
}
