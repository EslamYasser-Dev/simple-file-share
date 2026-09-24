import { API_BASE_URL } from '../config';
import { getToken } from './token';

export interface ServerEvent {
  type: string;
  path?: string;
  user?: string;
  at?: string;
}

type EventListener = (event: ServerEvent) => void;

let controller: AbortController | null = null;
let retryDelay = 1000;
let retryTimer: ReturnType<typeof setTimeout> | null = null;
let running = false;
const listeners = new Set<EventListener>();

const MAX_RETRY_MS = 30_000;

export function subscribeEvents(listener: EventListener): () => void {
  listeners.add(listener);
  return () => {
    listeners.delete(listener);
  };
}

function emit(event: ServerEvent): void {
  for (const listener of listeners) {
    try {
      listener(event);
    } catch {
      /* isolate listener errors */
    }
  }
}

function scheduleRetry(): void {
  if (!running) return;
  if (retryTimer) clearTimeout(retryTimer);
  retryTimer = setTimeout(() => {
    retryTimer = null;
    void openStream();
  }, retryDelay);
  retryDelay = Math.min(retryDelay * 2, MAX_RETRY_MS);
}

function parseSseChunk(buffer: string): { events: ServerEvent[]; rest: string } {
  const events: ServerEvent[] = [];
  const parts = buffer.split('\n\n');
  const rest = parts.pop() ?? '';
  for (const part of parts) {
    const lines = part.split('\n');
    let data = '';
    for (const line of lines) {
      if (line.startsWith(':')) continue;
      if (line.startsWith('data:')) {
        data += line.slice(5).trim();
      }
    }
    if (!data) continue;
    try {
      const parsed = JSON.parse(data) as ServerEvent;
      if (parsed && typeof parsed.type === 'string') events.push(parsed);
    } catch {
      /* ignore non-JSON frames */
    }
  }
  return { events, rest };
}

async function openStream(): Promise<void> {
  if (!running) return;
  const token = await getToken();
  if (!token) {
    scheduleRetry();
    return;
  }

  controller = new AbortController();
  let buffer = '';
  try {
    const response = await fetch(`${API_BASE_URL}/api/events`, {
      headers: { Authorization: `Bearer ${token}`, Accept: 'text/event-stream' },
      signal: controller.signal,
    });
    if (!response.ok) {
      controller = null;
      scheduleRetry();
      return;
    }
    retryDelay = 1000;
    const body = response.body;
    if (!body) {
      controller = null;
      scheduleRetry();
      return;
    }
    const reader = body.getReader();
    const decoder = new TextDecoder();
    for (;;) {
      const { done, value } = await reader.read();
      if (done) break;
      buffer += decoder.decode(value, { stream: true });
      const { events, rest } = parseSseChunk(buffer);
      buffer = rest;
      for (const event of events) emit(event);
    }
  } catch {
    /* aborted or network failure */
  } finally {
    controller = null;
    if (running) scheduleRetry();
  }
}

export function startEventStream(): void {
  if (running) return;
  running = true;
  retryDelay = 1000;
  void openStream();
}

export function stopEventStream(): void {
  running = false;
  if (retryTimer) {
    clearTimeout(retryTimer);
    retryTimer = null;
  }
  controller?.abort();
  controller = null;
}
