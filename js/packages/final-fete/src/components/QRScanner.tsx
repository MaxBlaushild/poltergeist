import { useEffect, useRef, useState, useCallback } from 'react';
import { Html5Qrcode } from 'html5-qrcode';
import { useNavigate } from 'react-router-dom';

export const QRScanner = () => {
  const navigate = useNavigate();
  const [error, setError] = useState<string | null>(null);
  const [scanning, setScanning] = useState(false);
  const scannerRef = useRef<Html5Qrcode | null>(null);
  const scannerId = 'qr-reader';

  const handleQRCodeDetected = useCallback((decodedText: string) => {
    // Stop scanning immediately to prevent multiple navigations
    if (scannerRef.current && scannerRef.current.isScanning) {
      scannerRef.current.stop().catch((err) => {
        console.error('Error stopping scanner:', err);
      });
    }

    // Extract room ID from URL
    // Expected format: https://domain.com/unlock-room/:roomId or /unlock-room/:roomId
    let roomId: string | null = null;

    try {
      const url = new URL(decodedText);
      const pathParts = url.pathname.split('/');
      const unlockIndex = pathParts.indexOf('unlock-room');
      if (unlockIndex !== -1 && pathParts[unlockIndex + 1]) {
        roomId = pathParts[unlockIndex + 1];
      }
    } catch {
      // If it's not a full URL, try parsing as relative path
      const pathParts = decodedText.split('/');
      const unlockIndex = pathParts.indexOf('unlock-room');
      if (unlockIndex !== -1 && pathParts[unlockIndex + 1]) {
        roomId = pathParts[unlockIndex + 1];
      } else {
        // If it's just a UUID, use it directly
        const uuidRegex = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
        if (uuidRegex.test(decodedText)) {
          roomId = decodedText;
        }
      }
    }

    if (roomId) {
      navigate(`/unlock-room/${roomId}`);
    } else {
      setError('Invalid QR code format. Expected URL with room ID.');
      // Restart scanning after error
      setTimeout(() => {
        if (scannerRef.current) {
          scannerRef.current.start(
            { facingMode: 'environment' },
            { fps: 10, qrbox: { width: 250, height: 250 } },
            handleQRCodeDetected,
            () => {}
          ).catch((err) => {
            console.error('Error restarting scanner:', err);
          });
        }
      }, 2000);
    }
  }, [navigate]);

  useEffect(() => {
    const html5QrCode = new Html5Qrcode(scannerId);
    scannerRef.current = html5QrCode;

    const startScanning = async () => {
      try {
        setScanning(true);
        setError(null);

        // Start scanning with camera
        await html5QrCode.start(
          { facingMode: 'environment' }, // Use back camera
          {
            fps: 10,
            qrbox: { width: 250, height: 250 },
          },
          (decodedText) => {
            // QR code detected
            handleQRCodeDetected(decodedText);
          },
          () => {
            // Ignore scanning errors (they're frequent during scanning)
          }
        );
      } catch (err: any) {
        console.error('Error starting QR scanner:', err);
        if (err.name === 'NotAllowedError' || err.name === 'PermissionDeniedError') {
          setError('Camera permission denied. Please allow camera access and try again.');
        } else if (err.name === 'NotFoundError') {
          setError('No camera found. Please ensure you have a camera connected.');
        } else {
          setError('Failed to start camera: ' + (err.message || 'Unknown error'));
        }
        setScanning(false);
      }
    };

    startScanning();

    // Cleanup on unmount
    return () => {
      if (scannerRef.current && scannerRef.current.isScanning) {
        scannerRef.current.stop().catch((err) => {
          console.error('Error stopping scanner:', err);
        });
      }
    };
  }, [handleQRCodeDetected]);

  const handleCancel = () => {
    if (scannerRef.current && scannerRef.current.isScanning) {
      scannerRef.current.stop().catch((err) => {
        console.error('Error stopping scanner:', err);
      });
    }
    navigate('/');
  };

  return (
    <div className="flex flex-col items-center justify-center min-h-screen p-2 md:p-4 bg-black">
      <h1 className="text-lg md:text-2xl font-bold text-[#00ff00] mb-4">Scan QR Code</h1>
      <p className="text-[#00ff00] mb-4 text-center px-4 opacity-80">
        Point your camera at the QR code to unlock a room
      </p>
      
      <div id={scannerId} className="w-full max-w-md mb-4 overflow-hidden border-2 border-[#00ff00] rounded-lg shadow-[0_0_20px_rgba(0,255,0,0.5)]"></div>
      
      {error && (
        <div className="bg-red-900/80 border-2 border-red-500 text-red-300 p-4 rounded-md mb-4 max-w-md text-center mx-4 shadow-[0_0_20px_rgba(255,0,0,0.5)]">
          {error}
        </div>
      )}

      {scanning && !error && (
        <p className="text-[#00ff00] mb-4">Scanning...</p>
      )}

      <button
        onClick={handleCancel}
        className="matrix-button matrix-button-secondary w-full md:w-auto min-h-[44px] max-w-md"
      >
        Cancel
      </button>
    </div>
  );
};

