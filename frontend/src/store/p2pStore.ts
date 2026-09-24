import { create } from 'zustand';

export interface P2PPeer {
  id: string;
  user?: string;
}

export interface P2PTransfer {
  id: string;
  peerId: string;
  peerLabel?: string;
  name: string;
  size: number;
  loaded: number;
  status: 'connecting' | 'active' | 'done' | 'error' | 'cancelled';
  direction: 'send' | 'receive';
  error?: string;
  /** Local blob URL for completed receives (revoke when cleared). */
  url?: string;
}

interface P2PState {
  peerId: string | null;
  connected: boolean;
  peers: P2PPeer[];
  transfers: P2PTransfer[];
  streamError: string | null;

  setPeerId: (id: string | null) => void;
  setConnected: (v: boolean) => void;
  setPeers: (peers: P2PPeer[]) => void;
  setStreamError: (msg: string | null) => void;
  upsertTransfer: (t: P2PTransfer) => void;
  patchTransfer: (id: string, patch: Partial<P2PTransfer>) => void;
  removeTransfer: (id: string) => void;
  clearTransfers: () => void;
}

function revokeUrl(url?: string) {
  if (!url) return;
  try {
    URL.revokeObjectURL(url);
  } catch {
    /* ignore */
  }
}

export const useP2PStore = create<P2PState>((set) => ({
  peerId: null,
  connected: false,
  peers: [],
  transfers: [],
  streamError: null,

  setPeerId: (peerId) => set({ peerId }),
  setConnected: (connected) => set({ connected }),
  setPeers: (peers) => set({ peers }),
  setStreamError: (streamError) => set({ streamError }),

  upsertTransfer: (t) =>
    set((state) => {
      const idx = state.transfers.findIndex((x) => x.id === t.id);
      if (idx === -1) return { transfers: [...state.transfers, t] };
      const next = [...state.transfers];
      next[idx] = { ...next[idx], ...t };
      return { transfers: next };
    }),

  patchTransfer: (id, patch) =>
    set((state) => ({
      transfers: state.transfers.map((x) => (x.id === id ? { ...x, ...patch } : x)),
    })),

  removeTransfer: (id) =>
    set((state) => {
      const target = state.transfers.find((x) => x.id === id);
      revokeUrl(target?.url);
      return { transfers: state.transfers.filter((x) => x.id !== id) };
    }),

  clearTransfers: () =>
    set((state) => {
      state.transfers.forEach((x) => revokeUrl(x.url));
      return { transfers: [] };
    }),
}));
