import { useEffect, useRef, useState, useCallback } from 'react';
import { Html5Qrcode } from 'html5-qrcode';
import axios from 'axios';

type ScanResult = {
  type: 'success' | 'error';
  message: string;
};

const errorMessage = 'ERROR: Attempted to inject antivirus software into a tutorial document.';

export const PressQRScanner = () => {
  const [error, setError] = useState<string | null>(null);
  const [scanning, setScanning] = useState(false);
  const [result, setResult] = useState<ScanResult | null>(null);
  const [processing, setProcessing] = useState(false);
  const scannerRef = useRef<Html5Qrcode | null>(null);
  const scannerId = 'press-qr-reader';

  const validatePressUrl = (urlString: string): { valid: boolean; slot?: number; error?: string } => {
    try {
      // Try to parse as full URL
      let url: URL;
      try {
        url = new URL(urlString);
      } catch {
        // If not a full URL, try to construct one with current origin
        url = new URL(urlString, window.location.origin);
      }

      const pathParts = url.pathname.split('/').filter(part => part);
      const pressIndex = pathParts.indexOf('press');
      
      if (pressIndex === -1 || pressIndex === pathParts.length - 1) {
        return { valid: false, error: errorMessage };
      }

      // Check if it's the utility closet puzzle press endpoint
      const utilityClosetIndex = pathParts.indexOf('utility-closet-puzzle');
      if (utilityClosetIndex === -1 || utilityClosetIndex !== pressIndex - 1) {
        return { valid: false, error: errorMessage };
      }

      const slotStr = pathParts[pressIndex + 1];
      const slot = parseInt(slotStr, 10);
      
      if (isNaN(slot) || slot < 0 || slot > 5) {
        return { valid: false, error: 'Invalid slot number. Slot must be between 0 and 5.' };
      }

      return { valid: true, slot };
    } catch (err) {
      return { valid: false, error: errorMessage };
    }
  };

  const makePressRequest = async (urlString: string): Promise<ScanResult> => {
    try {
      // Normalize URL - if it's a relative path, use current origin
      let fullUrl: string;
      try {
        new URL(urlString);
        fullUrl = urlString;
      } catch {
        fullUrl = `${window.location.origin}${urlString.startsWith('/') ? urlString : '/' + urlString}`;
      }

      // Always add antiviral=true query parameter to press URLs
      const url = new URL(fullUrl);
      url.searchParams.set('antiviral', 'true');
      fullUrl = url.toString();

      const response = await axios.get(fullUrl, {
        validateStatus: () => true, // Don't throw on any status code
      });

      if (response.status === 200) {
        return { type: 'success', message: 'Antiviral injected successfully!' };
      } else if (response.status === 403) {
        return { type: 'error', message: 'Antiviral required. Please visit /antiviral first to activate protection.' };
      } else if (response.status === 400) {
        const errorMsg = response.data?.error || 'Bad request';
        if (errorMsg.includes('slot') || errorMsg.includes('Slot')) {
          return { type: 'error', message: 'Invalid slot number. Slot must be between 0 and 5.' };
        }
        return { type: 'error', message: `Invalid request: ${errorMsg}` };
      } else if (response.status === 500) {
        const errorMsg = response.data?.error || 'Internal server error';
        return { type: 'error', message: `Server error: ${errorMsg}` };
      } else {
        return { type: 'error', message: `Request failed with status ${response.status}` };
      }
    } catch (err: any) {
      if (err.code === 'ERR_NETWORK' || err.message?.includes('Network Error')) {
        return { type: 'error', message: 'Failed to connect to server. Please check your internet connection.' };
      }
      return { type: 'error', message: `Error: ${err.message || 'Unknown error occurred'}` };
    }
  };

  const handleQRCodeDetected = useCallback(async (decodedText: string) => {
    // Stop scanning immediately
    if (scannerRef.current && scannerRef.current.isScanning) {
      try {
        await scannerRef.current.stop();
      } catch (err) {
        console.error('Error stopping scanner:', err);
      }
    }
    setScanning(false);
    setProcessing(true);
    setError(null);
    setResult(null);

    // Validate URL format
    const validation = validatePressUrl(decodedText);
    if (!validation.valid) {
      setResult({ type: 'error', message: validation.error || 'Invalid QR code format.' });
      setProcessing(false);
      return;
    }

    // Make the HTTP request
    const result = await makePressRequest(decodedText);
    setResult(result);
    setProcessing(false);
  }, []);

  const handleRescan = useCallback(async () => {
    setResult(null);
    setError(null);
    setProcessing(false);
    
    // Stop and clear the existing scanner if it exists
    if (scannerRef.current) {
      try {
        if (scannerRef.current.isScanning) {
          await scannerRef.current.stop();
        }
        scannerRef.current.clear();
      } catch (err) {
        console.error('Error clearing scanner:', err);
      }
    }
    
    // Create a new scanner instance
    const html5QrCode = new Html5Qrcode(scannerId);
    scannerRef.current = html5QrCode;
    
    try {
      setScanning(true);
      await html5QrCode.start(
        { facingMode: 'environment' },
        { fps: 10, qrbox: { width: 250, height: 250 } },
        handleQRCodeDetected,
        () => {}
      );
    } catch (err: any) {
      console.error('Error restarting scanner:', err);
      if (err.name === 'NotAllowedError' || err.name === 'PermissionDeniedError') {
        setError('Camera permission denied. Please allow camera access and try again.');
      } else if (err.name === 'NotFoundError') {
        setError('No camera found. Please ensure you have a camera connected.');
      } else {
        setError('Failed to start camera: ' + (err.message || 'Unknown error'));
      }
      setScanning(false);
    }
  }, [handleQRCodeDetected]);

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
      if (scannerRef.current) {
        if (scannerRef.current.isScanning) {
          scannerRef.current.stop().catch((err) => {
            console.error('Error stopping scanner:', err);
          });
        }
        try {
          scannerRef.current.clear();
        } catch (err) {
          console.error('Error clearing scanner:', err);
        }
      }
    };
  }, [handleQRCodeDetected]);

  return (
    <div className="flex flex-col items-center justify-center min-h-screen p-2 md:p-4 bg-black">
      <h1 className="text-lg md:text-2xl font-bold text-[#00ff00] mb-4">Scan Antiviral QR Code</h1>
      <p className="text-[#00ff00] mb-4 text-center px-4 opacity-80">
        Point your camera at the QR code to inject antiviral
      </p>
      
      <div 
        id={scannerId} 
        className={`w-full max-w-md mb-4 overflow-hidden border-2 border-[#00ff00] rounded-lg shadow-[0_0_20px_rgba(0,255,0,0.5)] ${result ? 'hidden' : ''}`}
      ></div>
      
      {error && (
        <div className="bg-red-900/80 border-2 border-red-500 text-red-300 p-4 rounded-md mb-4 max-w-md text-center mx-4 shadow-[0_0_20px_rgba(255,0,0,0.5)]">
          {error}
        </div>
      )}

      {processing && (
        <div className="bg-blue-900/80 border-2 border-blue-500 text-blue-300 p-4 rounded-md mb-4 max-w-md text-center mx-4 shadow-[0_0_20px_rgba(0,0,255,0.5)]">
          Processing request...
        </div>
      )}

      {result && (
        <div className={`border-2 p-4 rounded-md mb-4 max-w-md text-center mx-4 shadow-[0_0_20px_rgba(0,255,0,0.5)] ${
          result.type === 'success' 
            ? 'bg-green-900/80 border-green-500 text-green-300' 
            : 'bg-red-900/80 border-red-500 text-red-300'
        }`}>
          {result.message}
        </div>
      )}

      {scanning && !error && !result && (
        <p className="text-[#00ff00] mb-4">Scanning...</p>
      )}

      {result && (
        <button
          onClick={handleRescan}
          className="matrix-button matrix-button-secondary w-full md:w-auto min-h-[44px] max-w-md mb-2"
        >
          Scan Again
        </button>
      )}
    </div>
  );
};

