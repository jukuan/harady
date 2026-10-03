export const API_BASE = (import.meta.env.VITE_API_BASE as string | undefined) ?? 'http://localhost:8101'
export const WS_BASE  = (import.meta.env.VITE_WS_BASE  as string | undefined) ?? API_BASE.replace(/^http/, 'ws')

export const STORAGE_NICKNAME = 'harady:nickname'
