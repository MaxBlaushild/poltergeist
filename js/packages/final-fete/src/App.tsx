import React from 'react';
import { AuthProvider, APIProvider } from '@poltergeist/contexts';
import { createBrowserRouter, RouterProvider, Navigate } from 'react-router-dom';
import { RoomsList } from './components/RoomsList';
import { QRScanner } from './components/QRScanner';
import { UnlockRoom } from './components/UnlockRoom';
import { LoginPage } from './components/LoginPage';
import { MatrixBackground } from './components/MatrixBackground';
import { GarbledText } from './components/GarbledText';
import { AntiviralPage } from './components/AntiviralPage';
import { PressQRScanner } from './components/PressQRScanner';
import { MailRoomExtraLetterPage } from './components/MailRoomExtraLetterPage';
import { MailRoomNoteBreakerCluePage } from './components/MailRoomNoteBreakerCluePage';
import { WhiteIndicatorsPage } from './components/WhiteIndicatorsPage';
import { BlueIndicatorsPage } from './components/BlueIndicatorsPage';
import { PurpleIndicatorsPage } from './components/PurpleIndicatorsPage';
import { BlueRedPurplePage } from './components/BlueRedPurplePage';
import { WinningSequencePage } from './components/WinningSequencePage';
import { DataVaultPage } from './components/DataVaultPage';
import { SituationRoomPage } from './components/SituationRoomPage';
import { CodeEntryPage } from './components/CodeEntryPage';
import { MainLayout } from './components/MainLayout';
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
      path: '/antiviral',
      element: <AntiviralPage />,
    },
    {
      path: '/press-scanner',
      element: <PressQRScanner />,
    },
    {
      path: '/mail-room-extra-letter',
      element: <MailRoomExtraLetterPage />,
    },
    {
      path: '/mail-room-note-breaker-clue',
      element: <MailRoomNoteBreakerCluePage />,
    },
    {
      path: '/white-indicators',
      element: <WhiteIndicatorsPage />,
    },
    {
      path: '/blue-indicators',
      element: <BlueIndicatorsPage />,
    },
    {
      path: '/purple-indicators',
      element: <PurpleIndicatorsPage />,
    },
    {
      path: '/blue-red-purple',
      element: <BlueRedPurplePage />,
    },
    {
      path: '/winning-sequence',
      element: <WinningSequencePage />,
    },
    {
      path: '/data-vault',
      element: <DataVaultPage />,
    },
    {
      path: '/situation-room',
      element: <SituationRoomPage />,
    },
    {
      path: '/unlock-room/:roomId',
      element: <UnlockRoom />,
    },
    {
      path: '/garbled',
      element: (
        <ProtectedRoute>
          <GarbledText />
        </ProtectedRoute>
      ),
    },
    {
      element: (
        <ProtectedRoute>
          <MainLayout />
        </ProtectedRoute>
      ),
      children: [
        {
          path: '/',
          element: <RoomsList />,
        },
        {
          path: '/codes',
          element: <CodeEntryPage />,
        },
        {
          path: '/scan-qr',
          element: <QRScanner />,
        },
      ],
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
