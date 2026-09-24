import { buildUrl } from './api';
import { useFileStore } from '../store/fileStore';
import { useAuthStore } from '../store/authStore';
import api from './api';

export interface ServerEvent {
  type: string;
  path?: string;
  user?: string;
  at?: string;
}

let source: EventSource | null = null;
let retryTimer: ReturnType<typeof setTimeout> | null = null;
let retryDelay = 1000;
let refreshTimer: ReturnType<typeof setTimeout> | null = null;
let intentionallyClosed = false;

const MAX_RETRY_MS = 30_000;
const REFRESH_DEBOUNCE_MS = 150;

function scheduleRetry(): void {
  if (intentionallyClosed) return;
  if (retryTimer) clearTimeout(retryTimer);
  retryTimer = setTimeout(() => {
    retryTimer = null;
    openSource();
  }, retryDelay);
  retryDelay = Math.min(retryDelay * 2, MAX_RETRY_MS);
}

function refreshStores(): void {
  if (refreshTimer) clearTimeout(refreshTimer);
  refreshTimer = setTimeout(() => {
    refreshTimer = null;
    const files = useFileStore.getState();
    void files.fetchFiles(files.currentPath);
    void files.refreshRoot();
  }, REFRESH_DEBOUNCE_MS);
}

async function refreshQuota(): Promise<void> {
  const result = await api.me();
  if (result.data) {
    useAuthStore.getState().setUser(result.data);
  }
}

function handleEvent(data: ServerEvent): void {
  if (!data || typeof data.type !== 'string') return;
  if (data.type === 'quota') {
    void refreshQuota();
    return;
  }
  refreshStores();
}

function openSource(): void {
  if (typeof EventSource === 'undefined') return;
  if (source) {
    source.close();
    source = null;
  }
  intentionallyClosed = false;

  let es: EventSource;
  try {
    es = new EventSource(buildUrl('/api/events'), { withCredentials: true });
  } catch {
    scheduleRetry();
    return;
  }
  source = es;

  es.onopen = () => {
    retryDelay = 1000;
  };

  es.onmessage = (ev: MessageEvent<string>) => {
    try {
      const parsed = JSON.parse(ev.data) as ServerEvent;
      handleEvent(parsed);
    } catch {
      /* ignore malformed frames */
    }
  };

  es.onerror = () => {
    // EventSource retries automatically while CONNECTING; when CLOSED the
    // browser gave up (often after 401) — schedule our own backoff.
    if (es.readyState === EventSource.CLOSED) {
      es.close();
      if (source === es) source = null;
      // Probe session: a hard 401 should bounce the auth gate.
      void api
        .me()
        .then((result) => {
          if (result.unauthorized) {
            window.dispatchEvent(new CustomEvent('fs:unauthorized'));
            return;
          }
          scheduleRetry();
        })
        .catch(() => scheduleRetry());
    }
  };
}

/** Connect the live event stream (idempotent). */
export function startEvents(): void {
  if (source) return;
  openSource();
}

/** Disconnect the live event stream and cancel pending retries. */
export function stopEvents(): void {
  intentionallyClosed = true;
  if (retryTimer) {
    clearTimeout(retryTimer);
    retryTimer = null;
  }
  if (refreshTimer) {
    clearTimeout(refreshTimer);
    refreshTimer = null;
  }
  if (source) {
    source.close();
    source = null;
  }
  retryDelay = 1000;
}
