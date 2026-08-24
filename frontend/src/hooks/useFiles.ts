import { useContext } from 'react';
import { FileContext, type FileContextType } from '../contexts/fileContext';

export const useFiles = (): FileContextType => {
  const context = useContext(FileContext);
  if (context === undefined) {
    throw new Error('useFiles must be used within a FileProvider');
  }
  return context;
};