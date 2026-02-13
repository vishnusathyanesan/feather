import type { Channel } from "../types/channel";
import type { Message } from "../types/message";
import type { User } from "../types/user";

const STORAGE_KEY_PREFIX = "feather_cache_";

interface CacheData {
  channels: Channel[];
  messagesByChannel: Record<string, Message[]>;
  user: User | null;
  timestamp: number;
}

function getStorageKey(key: string): string {
  return `${STORAGE_KEY_PREFIX}${key}`;
}

export function saveToCache(data: Partial<CacheData>) {
  try {
    const existing = loadFromCache();
    const merged = { ...existing, ...data, timestamp: Date.now() };
    localStorage.setItem(getStorageKey("data"), JSON.stringify(merged));
  } catch {
    // Storage full or unavailable
  }
}

export function loadFromCache(): CacheData {
  try {
    const raw = localStorage.getItem(getStorageKey("data"));
    if (raw) {
      return JSON.parse(raw);
    }
  } catch {
    // Parse error
  }
  return {
    channels: [],
    messagesByChannel: {},
    user: null,
    timestamp: 0,
  };
}

export function cacheMessages(channelId: string, messages: Message[]) {
  const data = loadFromCache();
  data.messagesByChannel[channelId] = messages.slice(0, 100);
  saveToCache(data);
}

export function getCachedMessages(channelId: string): Message[] {
  const data = loadFromCache();
  return data.messagesByChannel[channelId] || [];
}

export function clearCache() {
  localStorage.removeItem(getStorageKey("data"));
}
