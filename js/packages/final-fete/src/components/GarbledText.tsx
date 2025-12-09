import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';

const chars = '01アイウエオカキクケコサシスセソタチツテトナニヌネノハヒフヘホマミムメモヤユヨラリルレロワヲンABCDEFGHIJKLMNOPQRSTUVWXYZ!@#$%^&*()_+-=[]{}|;:,.<>?';
const specialWords = ['laundry', 'stairs', 'ritual', 'portal', 'boiler', 'secretroom', 'itsokay', 'reallyitsok'];

export const GarbledText = () => {
  const navigate = useNavigate();
  const [text, setText] = useState('');
  const [isGenerating, setIsGenerating] = useState(true);

  useEffect(() => {
    // Generate garbled text
    const generateGarbledText = () => {
      const lines: string[] = [];
      const lineLength = 80;
      
      for (let i = 0; i < 50; i++) {
        let line = '';
        let j = 0;
        
        while (j < lineLength) {
          // 0.1% chance to insert a special word
          if (Math.random() < 0.03 && j < lineLength - 10) {
            const word = specialWords[Math.floor(Math.random() * specialWords.length)];
            // Check if word fits
            if (j + word.length <= lineLength) {
              line += word;
              j += word.length;
              // Add some garbled chars after the word
              const garbledAfter = Math.floor(Math.random() * 3);
              for (let k = 0; k < garbledAfter && j < lineLength; k++) {
                line += chars[Math.floor(Math.random() * chars.length)];
                j++;
              }
            } else {
              // Word doesn't fit, add regular char
              line += chars[Math.floor(Math.random() * chars.length)];
              j++;
            }
          } else {
            // Regular garbled character
            line += chars[Math.floor(Math.random() * chars.length)];
            j++;
          }
        }
        
        lines.push(line);
      }
      return lines.join('\n');
    };

    setText(generateGarbledText());
    setIsGenerating(false);
  }, []);

  return (
    <div className="min-h-screen bg-black text-[#00ff00] p-4 md:p-6 font-mono text-xs md:text-sm overflow-auto">
      <div className="max-w-4xl mx-auto">
        <button
          onClick={() => navigate('/')}
          className="mb-4 matrix-button matrix-button-secondary min-h-[44px]"
        >
          Back to Rooms
        </button>
        <pre className="whitespace-pre-wrap break-all" style={{ textShadow: '0 0 5px #00ff00' }}>
          {isGenerating ? 'LOADING...' : text}
        </pre>
      </div>
    </div>
  );
};

