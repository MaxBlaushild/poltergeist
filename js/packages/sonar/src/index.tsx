import React from 'react';
import ReactDOM from 'react-dom/client';
// import './index.css';
import App from './App.tsx';
import reportWebVitals from './reportWebVitals';

const rootElement = document.getElementById('root');
if (!rootElement) throw new Error('Failed to find the root element');
console.log(rootElement);
const root = ReactDOM.createRoot(rootElement);
root.render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
);

// Optional: Performance measuring
reportWebVitals(null);
