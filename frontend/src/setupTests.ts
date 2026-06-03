import "@testing-library/jest-dom";

// Mock ResizeObserver which is missing in jsdom
class ResizeObserverMock {
  observe() {}
  unobserve() {}
  disconnect() {}
}
globalThis.ResizeObserver = ResizeObserverMock;

// Mock window.scrollTo since it's not implemented in jsdom
window.scrollTo = () => {};
