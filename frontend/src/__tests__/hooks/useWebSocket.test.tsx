import { renderHook, act } from "@testing-library/react";
import { useWebSocket } from "@/hooks/useWebSocket";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import React from "react";

jest.mock("@/lib/api", () => ({
  getWsUrl: () => "ws://localhost:8080/ws/updates",
}));

class MockWebSocket {
  static instances: MockWebSocket[] = [];
  url: string;
  onmessage: ((e: MessageEvent) => void) | null = null;
  onerror: (() => void) | null = null;
  onclose: (() => void) | null = null;
  readyState = WebSocket.CONNECTING;

  constructor(url: string) {
    this.url = url;
    MockWebSocket.instances.push(this);
  }

  close() {
    this.readyState = WebSocket.CLOSED;
    if (this.onclose) this.onclose();
  }

  simulateMessage() {
    if (this.onmessage) this.onmessage(new MessageEvent("message", { data: "{}" }));
  }

  simulateError() {
    if (this.onerror) this.onerror();
  }
}

(global as unknown as { WebSocket: typeof MockWebSocket }).WebSocket = MockWebSocket;

function makeWrapper() {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  const Wrapper = ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={qc}>{children}</QueryClientProvider>
  );
  return { qc, Wrapper };
}

beforeEach(() => {
  MockWebSocket.instances = [];
  jest.useFakeTimers();
});

afterEach(() => {
  jest.useRealTimers();
});

describe("useWebSocket", () => {
  it("creates a WebSocket connection on mount", () => {
    const { Wrapper } = makeWrapper();
    renderHook(() => useWebSocket(), { wrapper: Wrapper });
    expect(MockWebSocket.instances).toHaveLength(1);
    expect(MockWebSocket.instances[0].url).toBe("ws://localhost:8080/ws/updates");
  });

  it("invalidates queries when a message is received", () => {
    const { qc, Wrapper } = makeWrapper();
    const invalidateSpy = jest.spyOn(qc, "invalidateQueries");
    renderHook(() => useWebSocket(), { wrapper: Wrapper });
    const ws = MockWebSocket.instances[0];

    act(() => {
      ws.simulateMessage();
    });

    expect(invalidateSpy).toHaveBeenCalledWith(
      expect.objectContaining({ queryKey: ["workflow-runs"] })
    );
    expect(invalidateSpy).toHaveBeenCalledWith(
      expect.objectContaining({ queryKey: ["task-runs"] })
    );
    expect(invalidateSpy).toHaveBeenCalledWith(
      expect.objectContaining({ queryKey: ["workers"] })
    );
    expect(invalidateSpy).toHaveBeenCalledWith(
      expect.objectContaining({ queryKey: ["workflows"] })
    );
  });

  it("closes the WebSocket on unmount", () => {
    const { Wrapper } = makeWrapper();
    const { unmount } = renderHook(() => useWebSocket(), { wrapper: Wrapper });
    const ws = MockWebSocket.instances[0];
    unmount();
    expect(ws.readyState).toBe(WebSocket.CLOSED);
  });

  it("schedules reconnect when connection closes", () => {
    const { Wrapper } = makeWrapper();
    renderHook(() => useWebSocket(), { wrapper: Wrapper });
    const ws = MockWebSocket.instances[0];

    act(() => {
      ws.onclose?.();
    });

    act(() => {
      jest.advanceTimersByTime(5000);
    });

    expect(MockWebSocket.instances).toHaveLength(2);
  });

  it("schedules reconnect on error", () => {
    const { Wrapper } = makeWrapper();
    renderHook(() => useWebSocket(), { wrapper: Wrapper });
    const ws = MockWebSocket.instances[0];

    act(() => {
      ws.simulateError();
    });

    act(() => {
      jest.advanceTimersByTime(5000);
    });

    expect(MockWebSocket.instances.length).toBeGreaterThanOrEqual(2);
  });
});
