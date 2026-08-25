/** Format a byte count into a human-readable string. */
export function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`;
}

/** Format an ISO date string into a localized date-time string. */
export function formatDate(iso: string): string {
  if (!iso) return '—';
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return '—';
  return d.toLocaleString(undefined, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}

/** Return a short relative time string (e.g. "2h ago"). */
export function timeAgo(iso: string): string {
  if (!iso) return '—';
  const date = new Date(iso);
  if (Number.isNaN(date.getTime())) return '—';
  const seconds = Math.floor((Date.now() - date.getTime()) / 1000);
  const units: [number, string][] = [
    [60, 's'],
    [60, 'm'],
    [24, 'h'],
    [7, 'd'],
    [4.35, 'w'],
    [12, 'mo'],
  ];
  let value = seconds;
  let unit = 's';
  for (const [div, u] of units) {
    if (value < div) break;
    value = Math.floor(value / div);
    unit = u;
  }
  return `${value}${unit} ago`;
}

/** Extract the file extension (lowercase, no dot) or empty string. */
export function getExtension(name: string): string {
  const idx = name.lastIndexOf('.');
  return idx === -1 ? '' : name.slice(idx + 1).toLowerCase();
}

/** Return a color key for a file extension (used for file-type badges). */
export function fileTypeColor(ext: string): string {
  const map: Record<string, string> = {
    pdf: 'bg-red-500/10 text-red-400 border-red-500/30',
    zip: 'bg-amber-500/10 text-amber-400 border-amber-500/30',
    rar: 'bg-amber-500/10 text-amber-400 border-amber-500/30',
    '7z': 'bg-amber-500/10 text-amber-400 border-amber-500/30',
    tar: 'bg-amber-500/10 text-amber-400 border-amber-500/30',
    gz: 'bg-amber-500/10 text-amber-400 border-amber-500/30',
    jpg: 'bg-emerald-500/10 text-emerald-400 border-emerald-500/30',
    jpeg: 'bg-emerald-500/10 text-emerald-400 border-emerald-500/30',
    png: 'bg-emerald-500/10 text-emerald-400 border-emerald-500/30',
    gif: 'bg-emerald-500/10 text-emerald-400 border-emerald-500/30',
    webp: 'bg-emerald-500/10 text-emerald-400 border-emerald-500/30',
    svg: 'bg-emerald-500/10 text-emerald-400 border-emerald-500/30',
    mp4: 'bg-purple-500/10 text-purple-400 border-purple-500/30',
    mov: 'bg-purple-500/10 text-purple-400 border-purple-500/30',
    avi: 'bg-purple-500/10 text-purple-400 border-purple-500/30',
    mp3: 'bg-pink-500/10 text-pink-400 border-pink-500/30',
    wav: 'bg-pink-500/10 text-pink-400 border-pink-500/30',
    ogg: 'bg-pink-500/10 text-pink-400 border-pink-500/30',
    txt: 'bg-sky-500/10 text-sky-400 border-sky-500/30',
    md: 'bg-sky-500/10 text-sky-400 border-sky-500/30',
    csv: 'bg-teal-500/10 text-teal-400 border-teal-500/30',
    json: 'bg-yellow-500/10 text-yellow-400 border-yellow-500/30',
    xml: 'bg-yellow-500/10 text-yellow-400 border-yellow-500/30',
    yaml: 'bg-yellow-500/10 text-yellow-400 border-yellow-500/30',
    yml: 'bg-yellow-500/10 text-yellow-400 border-yellow-500/30',
    exe: 'bg-slate-500/10 text-slate-400 border-slate-500/30',
    msi: 'bg-slate-500/10 text-slate-400 border-slate-500/30',
    deb: 'bg-slate-500/10 text-slate-400 border-slate-500/30',
    rpm: 'bg-slate-500/10 text-slate-400 border-slate-500/30',
    iso: 'bg-slate-500/10 text-slate-400 border-slate-500/30',
    img: 'bg-slate-500/10 text-slate-400 border-slate-500/30',
    doc: 'bg-blue-500/10 text-blue-400 border-blue-500/30',
    docx: 'bg-blue-500/10 text-blue-400 border-blue-500/30',
    xls: 'bg-green-500/10 text-green-400 border-green-500/30',
    xlsx: 'bg-green-500/10 text-green-400 border-green-500/30',
    ppt: 'bg-orange-500/10 text-orange-400 border-orange-500/30',
    pptx: 'bg-orange-500/10 text-orange-400 border-orange-500/30',
  };
  return map[ext] || 'bg-slate-500/10 text-slate-400 border-slate-500/30';
}