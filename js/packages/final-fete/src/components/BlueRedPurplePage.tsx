import React from 'react';

export const BlueRedPurplePage = () => {
  return (
    <div className="min-h-screen flex items-center justify-center p-4">
      <div className="bg-black/90 backdrop-blur-sm p-4 md:p-6 lg:p-8 rounded-lg border-2 border-[#00ff00] shadow-[0_0_20px_rgba(0,255,0,0.5)] w-full max-w-md mx-4 matrix-card">
        <h1 className="text-xl md:text-2xl font-bold mb-6 text-center text-[#00ff00]">Blue and Red Indicators</h1>
        <p className="text-[#00ff00] text-center text-lg md:text-xl">
          Applying antiviral to a server with a blue indicator next to red indicators will turn the red indicated servers purple.
        </p>
      </div>
    </div>
  );
};

