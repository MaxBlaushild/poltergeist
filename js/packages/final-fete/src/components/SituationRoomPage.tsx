import { useState, useRef } from 'react';
import type { KeyboardEvent } from 'react';

const CORRECT_CODE = ['7', '1', '3', '5', '4'];

export const SituationRoomPage = () => {
  const [inputs, setInputs] = useState<string[]>(['', '', '', '', '']);
  const [isSolved, setIsSolved] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const inputRefs = useRef<(HTMLInputElement | null)[]>([]);

  const handleInputChange = (index: number, value: string) => {
    // Only allow single character
    if (value.length > 1) return;
    
    const newInputs = [...inputs];
    newInputs[index] = value;
    setInputs(newInputs);
    // Clear error when user starts typing
    setError(null);

    // Auto-focus next input if a character was entered
    if (value && index < 4) {
      inputRefs.current[index + 1]?.focus();
    }
  };

  const handleKeyDown = (index: number, e: KeyboardEvent<HTMLInputElement>) => {
    // Handle backspace on empty input to focus previous
    if (e.key === 'Backspace' && !inputs[index] && index > 0) {
      inputRefs.current[index - 1]?.focus();
    }
  };

  const handleSubmit = () => {
    // Check if all inputs are filled
    if (inputs.some(input => !input)) {
      setError('Please enter all 5 digits');
      return;
    }

    // Check if inputs contain each correct character exactly once (order doesn't matter)
    const inputSorted = [...inputs].sort().join('');
    const correctSorted = [...CORRECT_CODE].sort().join('');
    
    if (inputSorted === correctSorted) {
      setIsSolved(true);
      setError(null);
    } else {
      setError('Incorrect combination. Please try again.');
    }
  };

  const handleReset = () => {
    setInputs(['', '', '', '', '']);
    setIsSolved(false);
    setError(null);
    inputRefs.current[0]?.focus();
  };

  if (isSolved) {
    return (
      <div className="min-h-screen flex items-center justify-center p-6">
        <div className="bg-black/85 border border-[#00ff00]/60 rounded-xl p-6 md:p-8 text-center shadow-[0_0_30px_rgba(0,255,0,0.3)] max-w-xl">
          <h1 className="text-2xl md:text-3xl font-bold mb-4 text-[#00ff00]">Situation Room</h1>
          <p className="text-sm md:text-base text-[#b5ffb5] leading-relaxed mb-6">
            Your prize is in the lockbox. The key is hidden under the white trash can.
          </p>
          <button
            onClick={handleReset}
            className="bg-[#00ff00]/20 hover:bg-[#00ff00]/30 border-2 border-[#00ff00] text-[#00ff00] font-bold py-3 px-6 rounded-lg transition-all duration-200 hover:shadow-[0_0_20px_rgba(0,255,0,0.5)]"
          >
            I have my prize
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen flex items-center justify-center p-6">
      <div className="bg-black/85 border border-[#00ff00]/60 rounded-xl p-6 md:p-8 shadow-[0_0_30px_rgba(0,255,0,0.3)] max-w-md w-full">
        <h1 className="text-2xl md:text-3xl font-bold mb-6 text-[#00ff00] text-center">Situation Room</h1>
        
        <div className="flex gap-3 justify-center mb-6">
          {inputs.map((value, index) => (
            <input
              key={index}
              ref={(el) => (inputRefs.current[index] = el)}
              type="text"
              inputMode="numeric"
              maxLength={1}
              value={value}
              onChange={(e) => handleInputChange(index, e.target.value)}
              onKeyDown={(e) => handleKeyDown(index, e)}
              className={`w-12 h-12 md:w-16 md:h-16 text-center text-2xl md:text-3xl font-bold bg-black/50 border-2 rounded-lg focus:outline-none focus:shadow-[0_0_15px_rgba(0,255,0,0.5)] ${
                error 
                  ? 'border-red-500/60 text-red-400 focus:border-red-500 focus:shadow-[0_0_15px_rgba(255,0,0,0.5)]' 
                  : 'border-[#00ff00]/60 text-[#00ff00] focus:border-[#00ff00]'
              }`}
            />
          ))}
        </div>

        {error && (
          <div className="mb-4 p-3 bg-red-500/10 border border-red-500/60 rounded-lg text-red-400 text-sm text-center">
            {error}
          </div>
        )}

        <button
          onClick={handleSubmit}
          className="w-full bg-[#00ff00]/20 hover:bg-[#00ff00]/30 border-2 border-[#00ff00] text-[#00ff00] font-bold py-3 px-6 rounded-lg transition-all duration-200 hover:shadow-[0_0_20px_rgba(0,255,0,0.5)]"
        >
          Submit
        </button>
      </div>
    </div>
  );
};

