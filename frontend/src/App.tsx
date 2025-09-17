import { useEffect, useRef, useState } from 'react'
import type { ChangeEvent } from 'react'

type UploadState = 'idle' | 'uploading' | 'success' | 'error'

// Vite provides typings via vite-env.d.ts, but ensure safe access in IDEs
const BASE_URL: string | undefined = (import.meta as any)?.env?.VITE_BASE_URL

function App() {
  const [status, setStatus] = useState<UploadState>('idle')
  const [message, setMessage] = useState<string>('')
  const [selectedCount, setSelectedCount] = useState<number>(0)
  const [targetPath, setTargetPath] = useState<string>('')
  const inputRef = useRef<HTMLInputElement | null>(null)

  useEffect(() => {
    // Enable folder selection (non-standard but widely supported)
    if (inputRef.current) {
      inputRef.current.setAttribute('webkitdirectory', '')
      inputRef.current.setAttribute('directory', '')
    }
  }, [])

  const onFilesChange = (e: ChangeEvent<HTMLInputElement>) => {
    const files = e.target.files as FileList | null
    setSelectedCount(files ? files.length : 0)
  }

  const upload = async () => {
    const el = inputRef.current
    if (!el || !el.files || el.files.length === 0) {
      setMessage('Please choose file(s) or a folder to upload.')
      setStatus('error')
      return
    }

    setStatus('uploading')
    setMessage('Uploading...')

    try {
      const form = new FormData()
      // Append files; backend iterates parts without requiring specific field name
      Array.from(el.files as FileList).forEach((f: File) => {
        const rel = (f as any).webkitRelativePath || f.name
        form.append('file', f, rel)
      })

      // Use proxy for /upload in dev; otherwise direct to BASE_URL
      const uploadUrl = BASE_URL ? `${BASE_URL}/upload` : '/upload'
      const res = await fetch(uploadUrl, {
        method: 'POST',
        body: form,
      })
      if (!res.ok) throw new Error(`Upload failed: ${res.status}`)
      const html = await res.text()
      setStatus('success')
      setMessage('Upload completed successfully.')
      // Open result page in a new tab for details
      const win = window.open('', '_blank')
      if (win) {
        win.document.write(html)
        win.document.close()
      }
    } catch (err: any) {
      console.error(err)
      setStatus('error')
      setMessage(err?.message || 'Upload failed')
    }
  }

  const openBrowser = () => {
    const url = (BASE_URL || '').trim() || window.location.origin
    window.open(`${url}/`, '_blank')
  }

  const openPath = () => {
    const url = (BASE_URL || '').trim() || window.location.origin
    const path = targetPath.startsWith('/') ? targetPath : `/${targetPath}`
    window.open(`${url}${path}`, '_blank')
  }

  const downloadZip = () => {
    const url = (BASE_URL || '').trim() || window.location.origin
    const path = targetPath.startsWith('/') ? targetPath : `/${targetPath}`
    window.open(`${url}${path}.zip`, '_blank')
  }

  return (
    <div className="mx-auto max-w-4xl p-6 text-gray-100">
      <h1 className="mb-2 text-2xl font-semibold">File Share Frontend</h1>
      <p className="mt-0 text-sm text-gray-400">Backend: {BASE_URL || 'Vite dev proxy (/upload)'}</p>

      <section className="mt-4 rounded-lg border border-gray-700 bg-gray-900/40 p-4">
        <h3 className="mb-3 text-lg font-medium">Upload Files or Folder</h3>
        <div className="flex flex-col gap-3 sm:flex-row sm:items-center">
          <input
            ref={inputRef}
            type="file"
            multiple
            onChange={onFilesChange}
            className="block w-full cursor-pointer rounded border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-200 file:mr-4 file:cursor-pointer file:rounded file:border-0 file:bg-indigo-600 file:px-3 file:py-2 file:text-sm file:font-medium file:text-white hover:file:bg-indigo-500"
          />
          <div className="text-sm text-gray-400">Selected: {selectedCount} item(s)</div>
        </div>
        <div className="mt-3">
          <button
            onClick={upload}
            disabled={status === 'uploading'}
            className="rounded bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-500 disabled:cursor-not-allowed disabled:opacity-60"
          >
            {status === 'uploading' ? 'Uploading...' : 'Upload'}
          </button>
        </div>
        {status !== 'idle' && (
          <div
            className={
              'mt-3 text-sm ' +
              (status === 'error' ? 'text-red-400' : 'text-emerald-400')
            }
          >
            {message}
          </div>
        )}
      </section>

      <section className="mt-4 rounded-lg border border-gray-700 bg-gray-900/40 p-4">
        <h3 className="mb-3 text-lg font-medium">Quick Actions</h3>
        <div className="mb-3 flex gap-2">
          <button
            onClick={openBrowser}
            className="rounded bg-slate-700 px-4 py-2 text-sm font-medium text-white hover:bg-slate-600"
          >
            Open File Browser
          </button>
        </div>
        <div className="flex items-center gap-2">
          <input
            type="text"
            placeholder="/path/in/storage (e.g. /docs)"
            value={targetPath}
            onChange={(e) => setTargetPath(e.target.value)}
            className="w-full rounded border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-200 placeholder:text-gray-500"
          />
          <button
            onClick={openPath}
            className="rounded bg-slate-700 px-4 py-2 text-sm font-medium text-white hover:bg-slate-600"
          >
            Open
          </button>
          <button
            onClick={downloadZip}
            className="rounded bg-slate-700 px-4 py-2 text-sm font-medium text-white hover:bg-slate-600"
          >
            Download .zip
          </button>
        </div>
      </section>
    </div>
  )
}

export default App
