import React from 'react';
import ReactDOM from 'react-dom/client';
import App from './App';
import WebApp from './WebApp';
import './style.css';

const isLocalWeb = window.location.protocol === 'http:' &&
  (window.location.hostname === '127.0.0.1' || window.location.hostname === 'localhost' || window.location.hostname === '[::1]');

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>{isLocalWeb ? <WebApp /> : <App />}</React.StrictMode>,
);
