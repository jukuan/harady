import * as Sentry from '@sentry/react'

const dsn = (import.meta.env.VITE_SENTRY_DSN as string | undefined)?.trim()

if (dsn) {
  Sentry.init({
    dsn,
    environment: import.meta.env.MODE,
    release: import.meta.env.VITE_APP_VERSION as string | undefined,
    enabled: import.meta.env.PROD,        // stay quiet in dev
    integrations: [
      Sentry.browserTracingIntegration(),
      Sentry.replayIntegration(),
    ],
    tracesSampleRate: import.meta.env.PROD ? 0.1 : 1.0,
    replaysSessionSampleRate: 0,
    replaysOnErrorSampleRate: 1.0,       // only replay sessions that error
    ignoreErrors: [
      'ResizeObserver loop limit exceeded',
      'WebSocket connection to',
    ],
  })
}
