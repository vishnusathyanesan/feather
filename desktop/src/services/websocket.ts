import type { WebSocketEvent, EventType } from "../types/websocket";
import { getAccessToken } from "./api";

type EventHandler = (event: WebSocketEvent) => void;

const WS_BASE = import.meta.env.VITE_WS_URL || "ws://localhost:8080/api/v1";

class WebSocketService {
  private ws: WebSocket | null = null;
  private handlers: Map<EventType, Set<EventHandler>> = new Map();
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 20;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private isIntentionallyClosed = false;

  connect() {
    const token = getAccessToken();
    if (!token) return;

    // Close existing connection before opening a new one
    if (this.ws) {
      this.ws.onclose = null;
      this.ws.close();
      this.ws = null;
    }

    this.isIntentionallyClosed = false;
    this.ws = new WebSocket(`${WS_BASE}/ws`);

    this.ws.onopen = () => {
      this.reconnectAttempts = 0;
      // Send auth as first message instead of token in URL
      this.ws?.send(JSON.stringify({
        type: "auth",
        payload: { token },
      }));
    };

    this.ws.onmessage = (event) => {
      try {
        const wsEvent: WebSocketEvent = JSON.parse(event.data);
        this.dispatch(wsEvent);
      } catch {
        // ignore malformed messages
      }
    };

    this.ws.onclose = () => {
      if (!this.isIntentionallyClosed) {
        this.scheduleReconnect();
      }
    };

    this.ws.onerror = () => {
      this.ws?.close();
    };
  }

  disconnect() {
    this.isIntentionallyClosed = true;
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
    this.ws?.close();
    this.ws = null;
  }

  on(type: EventType, handler: EventHandler) {
    if (!this.handlers.has(type)) {
      this.handlers.set(type, new Set());
    }
    this.handlers.get(type)!.add(handler);
    return () => this.off(type, handler);
  }

  off(type: EventType, handler: EventHandler) {
    this.handlers.get(type)?.delete(handler);
  }

  send(event: WebSocketEvent) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(event));
    }
  }

  sendTyping(channelId: string) {
    this.send({
      type: "typing",
      channel_id: channelId,
      payload: {},
    });
  }

  private dispatch(event: WebSocketEvent) {
    this.handlers.get(event.type)?.forEach((handler) => handler(event));
  }

  private scheduleReconnect() {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) return;

    const delay = Math.min(
      1000 * Math.pow(2, this.reconnectAttempts) + Math.random() * 1000,
      30000
    );
    this.reconnectAttempts++;

    this.reconnectTimer = setTimeout(() => {
      this.connect();
    }, delay);
  }
}

export const wsService = new WebSocketService();
