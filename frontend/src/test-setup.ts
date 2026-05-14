import '@testing-library/jest-dom/vitest'

// Mock ResizeObserver for antd components
globalThis.ResizeObserver = class ResizeObserver {
  observe() {}
  unobserve() {}
  disconnect() {}
}

// Mock matchMedia for antd responsive components
globalThis.matchMedia = globalThis.matchMedia || function() {
  return {
    matches: false,
    media: '',
    onchange: null,
    addListener: () => {},
    removeListener: () => {},
    addEventListener: () => {},
    removeEventListener: () => {},
    dispatchEvent: () => true,
  }
}
