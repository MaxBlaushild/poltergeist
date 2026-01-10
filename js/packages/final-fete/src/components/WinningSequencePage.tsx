import React from 'react';

const WINNING_SEQUENCE = [
  { slot: 0, color: 'Off (Gray)', value: 0 },
  { slot: 1, color: 'Blue', value: 1 },
  { slot: 2, color: 'Red', value: 2 },
  { slot: 3, color: 'White', value: 3 },
  { slot: 4, color: 'Yellow', value: 4 },
  { slot: 5, color: 'Purple', value: 5 },
];

export const WinningSequencePage = () => {
  return (
    <div className="min-h-screen flex items-center justify-center p-4">
      <div className="bg-black/90 backdrop-blur-sm p-4 md:p-6 lg:p-8 rounded-lg border-2 border-[#00ff00] shadow-[0_0_20px_rgba(0,255,0,0.5)] w-full max-w-2xl mx-4 matrix-card">
        <h1 className="text-xl md:text-2xl font-bold mb-6 text-center text-[#00ff00]">Server Room Virus Purge Sequence</h1>
        <div className="space-y-3">
          {WINNING_SEQUENCE.map((item) => (
            <div 
              key={item.slot} 
              className="flex items-center justify-between p-3 border border-[#00ff00]/50 rounded bg-black/50"
            >
              <span className="text-[#00ff00] font-semibold">{item.slot + 1}:</span>
              <span className="text-[#00ff00]">{item.color}</span>
            </div>
          ))}
        </div>
        <p className="text-[#00ff00]/80 text-center mt-6 text-sm">
          Success condition: All six slots must match this sequence exactly
        </p>
      </div>
    </div>
  );
};

