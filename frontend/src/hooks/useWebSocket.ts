"use client";

import { useEffect } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { getWsUrl } from "@/lib/api";

export function useWebSocket() {
  const queryClient = useQueryClient();

  useEffect(() => {
    let ws: WebSocket;
    let reconnectTimer: ReturnType<typeof setTimeout>;

    function connect() {
      try {
        ws = new WebSocket(getWsUrl());

        ws.onmessage = () => {
          queryClient.invalidateQueries({ queryKey: ["workflow-runs"] });
          queryClient.invalidateQueries({ queryKey: ["task-runs"] });
          queryClient.invalidateQueries({ queryKey: ["workers"] });
          queryClient.invalidateQueries({ queryKey: ["workflows"] });
        };

        ws.onerror = () => {
          ws.close();
        };

        ws.onclose = () => {
          reconnectTimer = setTimeout(connect, 5000);
        };
      } catch {
        reconnectTimer = setTimeout(connect, 5000);
      }
    }

    connect();

    return () => {
      clearTimeout(reconnectTimer);
      if (ws) ws.close();
    };
  }, [queryClient]);
}
