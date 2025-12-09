import React from 'react';
import { AuthProvider, APIProvider } from '@poltergeist/contexts';
import { createBrowserRouter, RouterProvider, Navigate } from 'react-router-dom';
import { RoomsList } from './components/RoomsList';
import { QRScanner } from './components/QRScanner';
import { UnlockRoom } from './components/UnlockRoom';
import { LoginPage } from './components/LoginPage';
import { MatrixBackground } from './components/MatrixBackground';
import { GarbledText } from './components/GarbledText';
import './App.css';

const ProtectedRoute = ({ children }: { children: React.ReactNode }) => {
  const token = localStorage.getItem('token');
  if (!token) {
    return <Navigate to="/login" replace />;
  }

  return <>{children}</>;
};

const OnlyUnauthenticatedRoute = ({ children }: { children: React.ReactNode }) => {
  const token = localStorage.getItem('token');
  if (token) {
    return <Navigate to="/" replace />;
  }

  return <>{children}</>;
};

function AppRouter() {
  const router = createBrowserRouter([
    {
      path: '/login',
      element: <OnlyUnauthenticatedRoute>
        <LoginPage />
      </OnlyUnauthenticatedRoute>,
    },
    {
      path: '/',
      element: (
        <ProtectedRoute>
          <RoomsList />
        </ProtectedRoute>
      ),
    },
    {
      path: '/scan-qr',
      element: (
        <ProtectedRoute>
          <QRScanner />
        </ProtectedRoute>
      ),
    },
    {
      path: '/unlock-room/:roomId',
      element: (
        <ProtectedRoute>
          <UnlockRoom />
        </ProtectedRoute>
      ),
    },
    {
      path: '/garbled',
      element: (
        <ProtectedRoute>
          <GarbledText />
        </ProtectedRoute>
      ),
    },
  ]);

  return <RouterProvider router={router} />;
}

function App() {
  return (
    <APIProvider>
      <AuthProvider appName="final-fete" uriPrefix="/final-fete">
        <div className="relative min-h-screen">
          <MatrixBackground />
          <div className="relative z-10">
            <AppRouter />
          </div>
        </div>
      </AuthProvider>
    </APIProvider>
  );
}

export default App;
