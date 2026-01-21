import React, { useRef } from 'react';
import { useNavigate } from 'react-router-dom';

export const MailRoomExtraLetterPage = () => {
  const navigate = useNavigate();
  const imgRef = useRef<HTMLImageElement>(null);

  const handleImageInteraction = (clientX: number, clientY: number) => {
    const img = imgRef.current;
    if (!img) return;

    const rect = img.getBoundingClientRect();
    const x = clientX - rect.left;
    const y = clientY - rect.top;

    // Calculate relative coordinates (0-1 range)
    const relX = x / rect.width;
    const relY = y / rect.height;

    // Check if touch/click is on a horizontal line area (red line)
    // Red line is typically horizontal, so check if click is within a horizontal band
    // Check bottom portion of image (70-100% from top) with wide horizontal tolerance (10-90% width)
    // This accounts for a red line that spans most of the width in the lower portion
    if (relY >= 0.70 && relY <= 1.0 && relX >= 0.10 && relX <= 0.90) {
      navigate('/garbled');
    }
  };

  const handleClick = (e: React.MouseEvent<HTMLImageElement>) => {
    handleImageInteraction(e.clientX, e.clientY);
  };

  const handleTouchStart = (e: React.TouchEvent<HTMLImageElement>) => {
    e.preventDefault(); // Prevent default touch behavior
    const touch = e.touches[0];
    if (touch) {
      handleImageInteraction(touch.clientX, touch.clientY);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center p-4 bg-black">
      <div className="bg-black/90 backdrop-blur-sm p-4 md:p-6 lg:p-8 rounded-lg border-2 border-[#00ff00] shadow-[0_0_20px_rgba(0,255,0,0.5)] w-full max-w-4xl mx-4 matrix-card">
        <div className="flex justify-center">
          <img 
            ref={imgRef}
            src="/Mail_Room_Extra_Letter_9b.webp" 
            alt="Mail Room Extra Letter" 
            className="max-w-full h-auto rounded-lg cursor-pointer"
            onClick={handleClick}
            onTouchStart={handleTouchStart}
          />
        </div>
      </div>
    </div>
  );
};

