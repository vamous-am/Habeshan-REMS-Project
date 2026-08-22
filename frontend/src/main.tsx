import React from 'react';
import ReactDOM from 'react-dom/client';
import App from './App';
import './styles/globals.css';
import { initStoragePersistence } from './lib/offline-db/storagePersist';

// Initialize browser persistent storage for Dexie/IndexedDB on startup
initStoragePersistence();

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
);