import '@testing-library/jest-dom/vitest'
import { afterEach } from 'vitest'
import { cleanup } from '@testing-library/react'

// ---------------------------------------------------------------------------
// jsdom polyfills.
//
// jsdom deliberately doesn't implement layout-related APIs. Any component
// that touches `scrollIntoView`, `matchMedia`, or an IntersectionObserver
// will throw unless we stub them here. These are no-ops, which is exactly
// what test code needs: they satisfy the call site without doing real work.
// ---------------------------------------------------------------------------

// scrollIntoView: called by GameScreen's mount-effect and per-append effect.
Element.prototype.scrollIntoView = function () {}

// matchMedia: used by Tailwind-driven `prefers-*` checks and some libraries.
if (!window.matchMedia) {
  window.matchMedia = (query: string) =>
    ({
      matches: false,
      media: query,
      onchange: null,
      addListener: () => {},
      removeListener: () => {},
      addEventListener: () => {},
      removeEventListener: () => {},
      dispatchEvent: () => false,
    }) as unknown as MediaQueryList
}

// IntersectionObserver: used by the Sentry Replay integration and any
// future infinite-scroll / lazy-load. Full spec is overkill — a stub that
// lets the call site succeed and stores the callback is enough for RTL.
class IO {
  private cb: IntersectionObserverCallback
  constructor(cb: IntersectionObserverCallback) {
    this.cb = cb
  }
  observe() {}
  unobserve() {}
  disconnect() {}
  takeRecords(): IntersectionObserverEntry[] {
    return []
  }
  readonly root = null
  readonly rootMargin = ''
  readonly thresholds: number[] = []
  // Expose the callback for future tests that want to trigger it manually.
  _trigger(entries: Partial<IntersectionObserverEntry>[]) {
    this.cb(entries as IntersectionObserverEntry[], this as unknown as IntersectionObserver)
  }
}
;(globalThis as unknown as { IntersectionObserver: unknown }).IntersectionObserver = IO

// ---------------------------------------------------------------------------
// Test cleanup.
// ---------------------------------------------------------------------------

afterEach(() => {
  cleanup()
})
