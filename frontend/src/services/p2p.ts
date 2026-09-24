import { buildUrl } from './api';
import { useP2PStore, type P2PPeer, type P2PTransfer } from '../store/p2pStore';

type SignalKind = 'offer' | 'answer' | 'candidate' | 'bye';

interface SignalFrame {
  type: 'signal';
  signal?: {
    from: string;
    to: string;
    kind: SignalKind;
    payload?: unknown;
    at?: string;
  };
}

interface HelloFrame {
  type: 'hello';
  peerId?: string;
  peers?: P2PPeer[];
}

interface PeersFrame {
  type: 'peers';
  peers?: P2PPeer[];
}

const ICE_SERVERS: RTCIceServer[] = [];
const CHUNK_BYTES = 16 * 1024;
const MAX_BUFFERED = 1_500_000;
const MAX_RETRY_MS = 30_000;

let source: EventSource | null = null;
let retryTimer: ReturnType<typeof setTimeout> | null = null;
let retryDelay = 1000;
let intentionallyClosed = false;

const pcs = new Map<string, RTCPeerConnection>();
const pendingCandidates = new Map<string, RTCIceCandidateInit[]>();
const incomingBuffers = new Map<
  string,
  { name: string; size: number; mime: string; peerId: string; chunks: BlobPart[] }
>();

function peerLabel(peerId: string): string | undefined {
  return useP2PStore.getState().peers.find((p) => p.id === peerId)?.user;
}

function newTransferId(): string {
  return `${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 10)}`;
}

async function sendSignal(to: string, kind: SignalKind, payload?: unknown): Promise<void> {
  const peerId = useP2PStore.getState().peerId;
  if (!peerId) throw new Error('no peer session');
  const res = await fetch(buildUrl('/api/p2p/signal'), {
    method: 'POST',
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      'X-Peer-Id': peerId,
    },
    body: JSON.stringify({ to, kind, payload }),
  });
  if (!res.ok) {
    const body = await res.text().catch(() => '');
    let message = `Signal failed (${res.status})`;
    try {
      const parsed = JSON.parse(body || '{}') as { error?: string };
      if (parsed.error) message = parsed.error;
    } catch {
      /* keep status message */
    }
    throw new Error(message);
  }
}

function createPeerConnection(remoteId: string): RTCPeerConnection {
  const existing = pcs.get(remoteId);
  if (existing) return existing;

  const pc = new RTCPeerConnection({ iceServers: ICE_SERVERS });
  pcs.set(remoteId, pc);

  pc.onicecandidate = (ev) => {
    if (!ev.candidate) return;
    void sendSignal(remoteId, 'candidate', ev.candidate.toJSON()).catch(() => undefined);
  };

  pc.onconnectionstatechange = () => {
    if (pc.connectionState === 'failed' || pc.connectionState === 'closed') {
      cleanupPeer(remoteId);
    }
  };

  pc.ondatachannel = (ev) => {
    attachDataChannel(remoteId, ev.channel);
  };

  const buffered = pendingCandidates.get(remoteId);
  if (buffered) {
    pendingCandidates.delete(remoteId);
    for (const c of buffered) {
      void pc.addIceCandidate(c).catch(() => undefined);
    }
  }

  return pc;
}

function cleanupPeer(remoteId: string) {
  const pc = pcs.get(remoteId);
  if (pc) {
    pcs.delete(remoteId);
    try {
      pc.close();
    } catch {
      /* ignore */
    }
  }
  pendingCandidates.delete(remoteId);
}

async function ensureRemoteDescription(
  pc: RTCPeerConnection,
  remoteId: string,
  desc: RTCSessionDescriptionInit,
): Promise<void> {
  await pc.setRemoteDescription(desc);
  const buffered = pendingCandidates.get(remoteId);
  if (buffered) {
    pendingCandidates.delete(remoteId);
    for (const c of buffered) {
      await pc.addIceCandidate(c).catch(() => undefined);
    }
  }
}

async function handleSignal(sig: NonNullable<SignalFrame['signal']>): Promise<void> {
  const { from, kind, payload } = sig;
  if (!from || from === useP2PStore.getState().peerId) return;

  if (kind === 'bye') {
    cleanupPeer(from);
    return;
  }

  const pc = createPeerConnection(from);

  if (kind === 'offer') {
    const desc = payload as RTCSessionDescriptionInit;
    await ensureRemoteDescription(pc, from, desc);
    const answer = await pc.createAnswer();
    await pc.setLocalDescription(answer);
    await sendSignal(from, 'answer', pc.localDescription?.toJSON() ?? answer);
    return;
  }

  if (kind === 'answer') {
    const desc = payload as RTCSessionDescriptionInit;
    if (pc.signalingState !== 'have-local-offer') {
      await ensureRemoteDescription(pc, from, desc);
      return;
    }
    await ensureRemoteDescription(pc, from, desc);
    return;
  }

  if (kind === 'candidate') {
    const cand = payload as RTCIceCandidateInit;
    if (!cand) return;
    if (!pc.remoteDescription) {
      const list = pendingCandidates.get(from) ?? [];
      list.push(cand);
      pendingCandidates.set(from, list);
      return;
    }
    await pc.addIceCandidate(cand).catch(() => undefined);
  }
}

function attachDataChannel(remoteId: string, channel: RTCDataChannel) {
  channel.binaryType = 'arraybuffer';

  channel.onmessage = (ev: MessageEvent) => {
    if (typeof ev.data === 'string') {
      try {
        const msg = JSON.parse(ev.data) as {
          t?: string;
          id?: string;
          name?: string;
          size?: number;
          mime?: string;
        };
        if (msg.t === 'meta' && msg.id && msg.name != null && typeof msg.size === 'number') {
          incomingBuffers.set(msg.id, {
            name: msg.name,
            size: msg.size,
            mime: msg.mime || 'application/octet-stream',
            peerId: remoteId,
            chunks: [],
          });
          const transfer: P2PTransfer = {
            id: msg.id,
            peerId: remoteId,
            peerLabel: peerLabel(remoteId),
            name: msg.name,
            size: msg.size,
            loaded: 0,
            status: 'active',
            direction: 'receive',
          };
          useP2PStore.getState().upsertTransfer(transfer);
          return;
        }
        if (msg.t === 'end' && msg.id) {
          const buf = incomingBuffers.get(msg.id);
          if (!buf) return;
          incomingBuffers.delete(msg.id);
          const blob = new Blob(buf.chunks, { type: buf.mime });
          const url = URL.createObjectURL(blob);
          useP2PStore.getState().patchTransfer(msg.id, {
            status: 'done',
            loaded: blob.size,
            url,
          });
          return;
        }
        if (msg.t === 'error' && msg.id) {
          incomingBuffers.delete(msg.id);
          useP2PStore.getState().patchTransfer(msg.id, {
            status: 'error',
            error: 'transfer failed',
          });
        }
      } catch {
        /* ignore malformed control frames */
      }
      return;
    }

    // Binary chunk: resolve which transfer by inflight single-stream assumption
    // plus size-tracking (meta opened the buffer; only one active receive per channel).
    const active = [...incomingBuffers.entries()];
    if (active.length === 0) return;
    // Prefer the buffer that still needs bytes.
    let entry = active[0];
    for (const candidate of active) {
      const need = candidate[1].size;
      const have = candidate[1].chunks.reduce((n, part) => n + (part as ArrayBuffer).byteLength, 0);
      if (have < need) {
        entry = candidate;
        break;
      }
    }
    const [id, buf] = entry;
    const data = ev.data as ArrayBuffer;
    buf.chunks.push(data);
    const loaded = buf.chunks.reduce((n, part) => n + (part as ArrayBuffer).byteLength, 0);
    useP2PStore.getState().patchTransfer(id, { loaded: Math.min(loaded, buf.size) });
  };
}

async function waitWritable(channel: RTCDataChannel): Promise<void> {
  if (channel.bufferedAmount <= MAX_BUFFERED) return;
  await new Promise<void>((resolve) => {
    const onLow = () => {
      channel.removeEventListener('bufferedamountlow', onLow);
      resolve();
    };
    channel.addEventListener('bufferedamountlow', onLow);
    const guard = setTimeout(() => {
      channel.removeEventListener('bufferedamountlow', onLow);
      resolve();
    }, 2000);
    void guard;
  });
}

async function streamFile(
  channel: RTCDataChannel,
  file: File,
  transferId: string,
): Promise<void> {
  const meta = JSON.stringify({
    t: 'meta',
    id: transferId,
    name: file.name,
    size: file.size,
    mime: file.type || 'application/octet-stream',
  });
  channel.send(meta);

  let offset = 0;
  while (offset < file.size) {
    const end = Math.min(offset + CHUNK_BYTES, file.size);
    const slice = file.slice(offset, end);
    const buf = await slice.arrayBuffer();
    await waitWritable(channel);
    if (channel.readyState !== 'open') {
      throw new Error('channel closed');
    }
    channel.send(buf);
    offset = end;
    useP2PStore.getState().patchTransfer(transferId, { loaded: offset });
  }

  channel.send(JSON.stringify({ t: 'end', id: transferId }));
}

/** Send one file peer-to-peer over a fresh WebRTC data channel. */
export async function sendFileToPeer(remoteId: string, file: File): Promise<string> {
  const transferId = newTransferId();
  const pc = createPeerConnection(remoteId);

  useP2PStore.getState().upsertTransfer({
    id: transferId,
    peerId: remoteId,
    peerLabel: peerLabel(remoteId),
    name: file.name,
    size: file.size,
    loaded: 0,
    status: 'connecting',
    direction: 'send',
  });

  try {
    const channel = pc.createDataChannel(`file-${transferId}`, { ordered: true });
    channel.binaryType = 'arraybuffer';
    channel.bufferedAmountLowThreshold = 256 * 1024;

    const opened = new Promise<void>((resolve, reject) => {
      const timer = setTimeout(() => reject(new Error('data channel timeout')), 20_000);
      channel.onopen = () => {
        clearTimeout(timer);
        resolve();
      };
      channel.onerror = () => {
        clearTimeout(timer);
        reject(new Error('data channel error'));
      };
    });

    const offer = await pc.createOffer();
    await pc.setLocalDescription(offer);
    await sendSignal(remoteId, 'offer', pc.localDescription?.toJSON() ?? offer);
    await opened;

    useP2PStore.getState().patchTransfer(transferId, { status: 'active' });
    await streamFile(channel, file, transferId);
    useP2PStore.getState().patchTransfer(transferId, {
      status: 'done',
      loaded: file.size,
    });
    return transferId;
  } catch (e) {
    const message = e instanceof Error ? e.message : 'send failed';
    useP2PStore.getState().patchTransfer(transferId, { status: 'error', error: message });
    throw e;
  }
}

function scheduleRetry(): void {
  if (intentionallyClosed) return;
  if (retryTimer) clearTimeout(retryTimer);
  retryTimer = setTimeout(() => {
    retryTimer = null;
    openStream();
  }, retryDelay);
  retryDelay = Math.min(retryDelay * 2, MAX_RETRY_MS);
}

function openStream(): void {
  if (typeof EventSource === 'undefined') return;
  if (source) {
    source.close();
    source = null;
  }
  intentionallyClosed = false;

  let es: EventSource;
  try {
    es = new EventSource(buildUrl('/api/p2p/stream'), { withCredentials: true });
  } catch {
    scheduleRetry();
    return;
  }
  source = es;

  es.addEventListener('hello', (ev) => {
    try {
      const frame = JSON.parse((ev as MessageEvent<string>).data) as HelloFrame;
      retryDelay = 1000;
      useP2PStore.getState().setConnected(true);
      useP2PStore.getState().setStreamError(null);
      if (frame.peerId) useP2PStore.getState().setPeerId(frame.peerId);
      if (frame.peers) useP2PStore.getState().setPeers(frame.peers);
    } catch {
      /* ignore */
    }
  });

  es.addEventListener('peers', (ev) => {
    try {
      const frame = JSON.parse((ev as MessageEvent<string>).data) as PeersFrame;
      useP2PStore.getState().setPeers(frame.peers ?? []);
    } catch {
      /* ignore */
    }
  });

  es.addEventListener('signal', (ev) => {
    try {
      const frame = JSON.parse((ev as MessageEvent<string>).data) as SignalFrame;
      if (frame.signal) void handleSignal(frame.signal).catch(() => undefined);
    } catch {
      /* ignore */
    }
  });

  es.onmessage = () => {
    /* untyped frames unused */
  };

  es.onerror = () => {
    useP2PStore.getState().setConnected(false);
    if (es.readyState === EventSource.CLOSED) {
      es.close();
      if (source === es) source = null;
      useP2PStore.getState().setStreamError('reconnecting');
      scheduleRetry();
    }
  };
}

export function startP2P(): void {
  if (source) return;
  openStream();
}

export function stopP2P(): void {
  intentionallyClosed = true;
  if (retryTimer) {
    clearTimeout(retryTimer);
    retryTimer = null;
  }
  if (source) {
    source.close();
    source = null;
  }
  for (const id of [...pcs.keys()]) {
    try {
      const pc = pcs.get(id);
      if (pc && useP2PStore.getState().peerId) {
        void sendSignal(id, 'bye', undefined).catch(() => undefined);
      }
    } catch {
      /* ignore */
    }
    cleanupPeer(id);
  }
  incomingBuffers.clear();
  retryDelay = 1000;
  useP2PStore.getState().setConnected(false);
  useP2PStore.getState().setPeerId(null);
  useP2PStore.getState().setPeers([]);
}
