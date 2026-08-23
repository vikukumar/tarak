"use client";

import { useEffect, useRef, useState, useCallback } from "react";
import { getAuthToken } from "@/lib/api";

export interface WsEvent {
  type: string;
  namespace?: string;
  resource?: string;
  data?: any;
  timestamp: string;
}

export function useTarakWebSocket(onEvent?: (evt: WsEvent) => void) {
  const [isConnected, setIsConnected] = useState(false);
  const [lastEvent, setLastEvent] = useState<WsEvent | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);

  const connect = useCallback(() => {
    if (typeof window === "undefined") return;

    // Clean up existing
    if (wsRef.current) {
      wsRef.current.close();
    }

    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    const host = window.location.host;
    const token = getAuthToken();
    const url = `${protocol}//${host}/apis/ws.tarak.io/v1/live${token ? `?token=${token}` : ""}`;

    try {
      const ws = new WebSocket(url);
      wsRef.current = ws;

      ws.onopen = () => {
        setIsConnected(true);
      };

      ws.onmessage = (event) => {
        try {
          const parsed: WsEvent = JSON.parse(event.data);
          setLastEvent(parsed);
          if (onEvent) {
            onEvent(parsed);
          }
        } catch {
          // ignore unparsable
        }
      };

      ws.onclose = () => {
        setIsConnected(false);
        // Automatic reconnect with backoff
        reconnectTimeoutRef.current = setTimeout(connect, 3000);
      };

      ws.onerror = () => {
        setIsConnected(false);
      };
    } catch {
      setIsConnected(false);
      reconnectTimeoutRef.current = setTimeout(connect, 5000);
    }
  }, [onEvent]);

  useEffect(() => {
    connect();
    return () => {
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
  }, [connect]);

  return { isConnected, lastEvent };
}
