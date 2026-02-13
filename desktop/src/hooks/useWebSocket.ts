import { useEffect, useRef } from "react";
import { wsService } from "../services/websocket";
import type { EventType, WebSocketEvent } from "../types/websocket";

export function useWebSocketEvent(
  type: EventType,
  handler: (event: WebSocketEvent) => void
) {
  const handlerRef = useRef(handler);
  handlerRef.current = handler;

  useEffect(() => {
    const unsub = wsService.on(type, (event) => handlerRef.current(event));
    return unsub;
  }, [type]);
}
