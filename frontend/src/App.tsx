import { useRef, useState, useCallback, useEffect } from 'react';
import type { ReactNode } from 'react';
import { FileProvider, useFiles } from './contexts';
import { useFileOperations } from './hooks';
import type { FileItem } from './services/api';
import './App.css';

// Main App component wrapped with FileProvider
function App() {
  return (
    <FileProvider>
      <AppContent />
    </FileProvider>
  );
}

// Inner component that can use the FileContext
function AppContent(): ReactNode {
  const { 
    files, 
    currentPath, 
    isLoading, 
    error, 
    uploadFile, 
    fetchFiles, 
    navigateToPath,
    clearError 
  } = useFiles();
  
  const { uploadProgress } = useFileOperations();
  const [selectedFiles, setSelectedFiles] = useState<FileList | null>(null);
  const [uploadStatus, setUploadStatus] = useState<'idle' | 'uploading' | 'success' | 'error'>('idle');
  const [message, setMessage] = useState('');
  const fileInputRef = useRef<HTMLInputElement>(null);

  // Fetch files when the component mounts or currentPath changes
  useEffect(() => {
    fetchFiles(currentPath);
  }, [currentPath, fetchFiles]);

  // Handle file selection
  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files.length > 0) {
      setSelectedFiles(e.target.files);
      setUploadStatus('idle');
      setMessage('');
    }
  };

  // Handle file upload
  const handleUpload = useCallback(async () => {
    if (!selectedFiles || selectedFiles.length === 0) {
      setMessage('Please select at least one file to upload');
      setUploadStatus('error');
      return;
    }

    setUploadStatus('uploading');
    setMessage('Uploading files...');

    try {
      const results = await Promise.all(
        Array.from(selectedFiles).map(file => uploadFile(file, currentPath))
      );

      const successCount = results.filter(r => !r.error).length;
      const errorCount = results.length - successCount;
      
      if (errorCount > 0) {
        setMessage(`Upload complete with ${errorCount} error(s)`);
        setUploadStatus('error');
      } else {
        setMessage('All files uploaded successfully!');
        setUploadStatus('success');
      }
      
      // Refresh the file list
      await fetchFiles(currentPath);
      
      // Clear the file input
      if (fileInputRef.current) {
        fileInputRef.current.value = '';
      }
      setSelectedFiles(null);
      
    } catch (err) {
      setMessage(`Upload failed: ${err instanceof Error ? err.message : 'Unknown error'}`);
      setUploadStatus('error');
    }
  }, [selectedFiles, currentPath, uploadFile, fetchFiles]);

  // Handle navigation to a directory
  const handleNavigate = useCallback((path: string) => {
    navigateToPath(path);
    setUploadStatus('idle');
    setMessage('');
  }, [navigateToPath]);

  // Handle going up one directory
  const handleGoUp = useCallback(() => {
    if (!currentPath) return;
    const pathParts = currentPath.split('/');
    pathParts.pop();
    const parentPath = pathParts.join('/');
    navigateToPath(parentPath);
    setUploadStatus('idle');
    setMessage('');
  }, [currentPath, navigateToPath]);

  // Handle file or directory click
  const handleItemClick = useCallback((item: FileItem) => {
    if (item.isDir) {
      handleNavigate(item.path);
    } else {
      // Handle file click (e.g., preview or download)
      try {
        // Get the base URL from environment variables or use current origin
        const baseUrl = import.meta.env.VITE_API_BASE_URL || window.location.origin;
        // Create a proper URL object to handle path construction
        const url = new URL('/api/files/download', baseUrl);
        // Add the path parameter
        url.searchParams.append('path', item.path);
        // Open the URL in a new tab
        window.open(url.toString(), '_blank');
      } catch (error) {
        console.error('Error constructing download URL:', error);
        setMessage(`Error downloading file: ${error instanceof Error ? error.message : 'Unknown error'}`);
        setUploadStatus('error');
      }
    }
  }, [handleNavigate]);

  // Render the file list
  const renderFileList = () => {
    if (isLoading) {
      return (
        <div className="text-center py-12">
          <div className="inline-block animate-spin rounded-full h-8 w-8 border-4 border-indigo-500 border-t-transparent"></div>
          <p className="mt-2 text-gray-500">Loading...</p>
        </div>
      );
    }

    if (error) {
      return (
        <div className="bg-red-50 border-l-4 border-red-400 p-4 rounded">
          <div className="flex">
            <div className="flex-shrink-0">
              <svg className="h-5 w-5 text-red-400" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
                <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
              </svg>
            </div>
            <div className="ml-3">
              <p className="text-sm text-red-700">
                Error: {error}
              </p>
              <div className="mt-2">
                <button
                  onClick={() => clearError()}
                  className="text-sm font-medium text-red-700 hover:text-red-600 transition-colors"
                >
                  Dismiss
                </button>
              </div>
            </div>
          </div>
        </div>
      );
    }

    if (!files || files.length === 0) {
      return (
        <div className="text-center py-12">
          <svg
            className="mx-auto h-12 w-12 text-gray-400"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            aria-hidden="true"
          >
            <path
              vectorEffect="non-scaling-stroke"
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M9 13h6m-3-3v6m-9 1V7a2 2 0 012-2h6l2 2h6a2 2 0 012 2v8a2 2 0 01-2 2H5a2 2 0 01-2-2z"
            />
          </svg>
          <h3 className="mt-2 text-sm font-medium text-gray-900">No files</h3>
          <p className="mt-1 text-sm text-gray-500">
            Get started by uploading a new file.
          </p>
        </div>
      );
    }

    return (
      <div className="bg-white shadow overflow-hidden sm:rounded-md">
        <ul className="divide-y divide-gray-200">
          {files.map((file) => (
            <li key={file.path}>
              <div 
                onClick={() => handleItemClick(file)}
                className={`block hover:bg-gray-50 cursor-pointer ${
                  file.isDir ? 'bg-indigo-50' : 'bg-white'
                }`}
              >
                <div className="px-4 py-4 sm:px-6">
                  <div className="flex items-center justify-between">
                    <p className={`text-sm font-medium ${
                      file.isDir ? 'text-indigo-600' : 'text-gray-700'
                    } truncate`}>
                      {file.name}
                    </p>
                    <div className="ml-2 flex-shrink-0 flex">
                      {!file.isDir && (
                        <p className="px-2 inline-flex text-xs leading-5 font-semibold rounded-full bg-green-100 text-green-800">
                          {(file.size / 1024).toFixed(2)} KB
                        </p>
                      )}
                    </div>
                  </div>
                  <div className="mt-2 sm:flex sm:justify-between">
                    <div className="sm:flex">
                      <p className="flex items-center text-sm text-gray-500">
                        {file.isDir ? (
                          <>
                            <svg className="flex-shrink-0 mr-1.5 h-5 w-5 text-gray-400" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
                              <path fillRule="evenodd" d="M2 6a2 2 0 012-2h4l2 2h4a2 2 0 012 2v8a2 2 0 01-2 2H4a2 2 0 01-2-2V6z" clipRule="evenodd" />
                            </svg>
                            Directory
                          </>
                        ) : (
                          <>
                            <svg className="flex-shrink-0 mr-1.5 h-5 w-5 text-gray-400" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
                              <path fillRule="evenodd" d="M4 4a2 2 0 012-2h4.586A2 2 0 0112 2.586L15.414 6A2 2 0 0116 7.414V16a2 2 0 01-2 2H6a2 2 0 01-2-2V4z" clipRule="evenodd" />
                            </svg>
                            File
                          </>
                        )}
                      </p>
                    </div>
                    <div className="mt-2 flex items-center text-sm text-gray-500 sm:mt-0">
                      <svg className="flex-shrink-0 mr-1.5 h-5 w-5 text-gray-400" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
                        <path fillRule="evenodd" d="M6 2a1 1 0 00-1 1v1H4a2 2 0 00-2 2v10a2 2 0 002 2h12a2 2 0 002-2V6a2 2 0 00-2-2h-1V3a1 1 0 10-2 0v1H7V3a1 1 0 00-1-1zm0 5a1 1 0 000 2h8a1 1 0 100-2H6z" clipRule="evenodd" />
                      </svg>
                      <p>
                        <time dateTime={file.modified}>
                          {new Date(file.modified).toLocaleString()}
                        </time>
                      </p>
                    </div>
                  </div>
                </div>
              </div>
            </li>
          ))}
        </ul>
      </div>
    );
  };

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white shadow-sm">
        <div className="max-w-7xl mx-auto px-4 py-4 sm:px-6 lg:px-8">
          <h1 className="text-3xl font-bold text-indigo-600">File Share</h1>
          <div className="mt-4 flex flex-wrap items-center gap-4">
            <input
              type="file"
              ref={fileInputRef}
              onChange={handleFileChange}
              multiple
              className="hidden"
            />
            <button
              onClick={() => fileInputRef.current?.click()}
              disabled={uploadStatus === 'uploading'}
              className="px-4 py-2 bg-indigo-600 text-white rounded-md hover:bg-indigo-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              Select Files
            </button>
            <button
              onClick={handleUpload}
              disabled={!selectedFiles || selectedFiles.length === 0 || uploadStatus === 'uploading'}
              className="px-4 py-2 bg-green-600 text-white rounded-md hover:bg-green-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              {uploadStatus === 'uploading' ? 'Uploading...' : 'Upload'}
            </button>
            {uploadStatus === 'uploading' && (
              <div className="flex-1 max-w-md h-2 bg-gray-200 rounded-full overflow-hidden">
                <div
                  className="h-full bg-indigo-600 transition-all duration-300"
                  style={{ width: `${uploadProgress}%` }}
                />
              </div>
            )}
          </div>
        </div>
      </header>

      <div className="max-w-7xl mx-auto px-4 py-4 sm:px-6 lg:px-8 flex items-center gap-2 border-b border-gray-200">
        <button
          onClick={handleGoUp}
          disabled={!currentPath}
          className={`px-3 py-1 text-sm rounded ${currentPath ? 'text-gray-700 hover:bg-gray-100' : 'text-gray-400 cursor-not-allowed'}`}
        >
          Up
        </button>
        <span className="text-sm text-gray-500 font-mono truncate">
          {currentPath || '/'}
        </span>
      </div>

      {message && (
        <div className={`max-w-7xl mx-auto px-4 py-3 sm:px-6 lg:px-8 rounded-md ${
          uploadStatus === 'error' ? 'bg-red-50 text-red-700' : 
          uploadStatus === 'success' ? 'bg-green-50 text-green-700' : 
          'bg-blue-50 text-blue-700'
        }`}>
          <div className="flex justify-between items-center">
            <p>{message}</p>
            <button 
              onClick={() => setMessage('')}
              className="text-xl leading-none opacity-70 hover:opacity-100"
            >
              &times;
            </button>
          </div>
        </div>
      )}

      <main className="max-w-7xl mx-auto px-4 py-6 sm:px-6 lg:px-8">
        {renderFileList()}
      </main>

      <footer className="border-t border-gray-200 mt-8">
        <div className="max-w-7xl mx-auto px-4 py-6 sm:px-6 lg:px-8">
          <p className="text-center text-gray-500 text-sm">
            © {new Date().getFullYear()} File Share App
          </p>
        </div>
      </footer>
    </div>
  );
}

export default App;
