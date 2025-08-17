import React, { useEffect, useState } from 'react';

export const Loader = () => {
    const [dots, setDots] = useState('');
  
    useEffect(() => {
      const interval = setInterval(() => {
        setDots(prev => prev.length >= 3 ? '' : prev + '.');
      }, 1000);
  
      return () => clearInterval(interval);
    }, []);
  
    return (
      <div className="flex flex-col items-center justify-center h-screen">
        <img
          src="/pirate-ship.png"
          alt="Pirate Ship"
          className="Layout__icon animate-spin w-32 h-32 mb-4"
        />
        <div className="text-xl font-semibold">
          Loading{dots}
        </div>
      </div>
    );
  };