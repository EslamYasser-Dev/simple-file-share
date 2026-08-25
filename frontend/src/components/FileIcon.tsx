import {
  File,
  FileArchive,
  FileAudio,
  FileCode,
  FileImage,
  FileSpreadsheet,
  FileText,
  FileVideo,
  Folder,
  type LucideIcon,
} from 'lucide-react';
import { getExtension } from '../lib/utils';

const IMAGE_EXT = new Set(['jpg', 'jpeg', 'png', 'gif', 'webp', 'svg', 'bmp', 'ico']);
const VIDEO_EXT = new Set(['mp4', 'mov', 'avi', 'mkv', 'webm', 'flv', 'wmv']);
const AUDIO_EXT = new Set(['mp3', 'wav', 'ogg', 'flac', 'aac', 'm4a']);
const ARCHIVE_EXT = new Set(['zip', 'rar', '7z', 'tar', 'gz', 'bz2', 'xz']);
const CODE_EXT = new Set(['js', 'ts', 'tsx', 'jsx', 'go', 'py', 'java', 'c', 'cpp', 'h', 'hpp', 'cs', 'rb', 'php', 'sh', 'bash', 'zsh', 'sql', 'html', 'css', 'scss', 'sass', 'less', 'json', 'yaml', 'yml', 'toml', 'ini', 'cfg', 'conf', 'env', 'gitignore', 'dockerfile', 'makefile']);
const SPREADSHEET_EXT = new Set(['xls', 'xlsx', 'csv', 'tsv', 'ods']);
const TEXT_EXT = new Set(['txt', 'md', 'log', 'rtf', 'pdf', 'doc', 'docx', 'odt']);

interface FileIconProps {
  name: string;
  isDir: boolean;
  className?: string;
}

export function FileIcon({ name, isDir, className = 'h-5 w-5' }: FileIconProps) {
  if (isDir) {
    return <Folder className={`${className} text-amber-400`} />;
  }

  const ext = getExtension(name);
  let Icon: LucideIcon = File;

  if (IMAGE_EXT.has(ext)) Icon = FileImage;
  else if (VIDEO_EXT.has(ext)) Icon = FileVideo;
  else if (AUDIO_EXT.has(ext)) Icon = FileAudio;
  else if (ARCHIVE_EXT.has(ext)) Icon = FileArchive;
  else if (CODE_EXT.has(ext)) Icon = FileCode;
  else if (SPREADSHEET_EXT.has(ext)) Icon = FileSpreadsheet;
  else if (TEXT_EXT.has(ext)) Icon = FileText;

  const color = ext
    ? 'text-slate-400'
    : 'text-slate-500';

  return <Icon className={`${className} ${color}`} />;
}