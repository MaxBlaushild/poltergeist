import React from 'react';
import ReactDOM from 'react-dom/client';
import './index.css';
import './output.css';
import 'mapbox-gl/dist/mapbox-gl.css';
import App from './App.tsx';
import mapboxgl from 'mapbox-gl';

const rootElement = document.getElementById('root');
if (!rootElement) throw new Error('Failed to find the root element');
const root = ReactDOM.createRoot(rootElement);
root.render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
);
