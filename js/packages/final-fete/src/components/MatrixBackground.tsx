import { useEffect, useRef, useState } from 'react';

export const MatrixBackground = () => {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const pulsingColumnRef = useRef<number | null>(null);
  const pulsePhaseRef = useRef(0);
  const isPulsingRef = useRef(false);
  const [isPulsing, setIsPulsing] = useState(false); // Only for cursor style

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    // Set canvas size
    const resizeCanvas = () => {
      canvas.width = window.innerWidth;
      canvas.height = window.innerHeight;
    };
    resizeCanvas();
    window.addEventListener('resize', resizeCanvas);

    // Matrix characters - mix of alphanumeric and symbols
    const chars = '01アイウエオカキクケコサシスセソタチツテトナニヌネノハヒフヘホマミムメモヤユヨラリルレロワヲン';
    const charArray = chars.split('');

    // Column configuration
    const fontSize = 14;
    const columns = Math.floor(canvas.width / fontSize);
    
    // Adjust for mobile performance
    const isMobile = window.innerWidth < 768;
    const activeColumns = isMobile ? Math.floor(columns * 0.6) : columns;
    
    // Create drops array - one per column
    const drops: number[] = [];
    for (let i = 0; i < activeColumns; i++) {
      drops[i] = Math.random() * -100; // Random starting position
    }

    // Select a random column to pulse red
    const selectPulsingColumn = () => {
      const randomColumn = Math.floor(Math.random() * activeColumns);
      pulsingColumnRef.current = randomColumn;
      isPulsingRef.current = true;
      pulsePhaseRef.current = 0;
      setIsPulsing(true);
    };

    // Initial selection
    selectPulsingColumn();

    // Select new pulsing column every 30 seconds
    const pulseInterval = setInterval(() => {
      selectPulsingColumn();
    }, 20000);

    // Pulse animation (5 seconds = 5000ms, at 35ms per frame = ~143 frames)
    const pulseAnimation = setInterval(() => {
      if (isPulsingRef.current) {
        pulsePhaseRef.current += 1;
        if (pulsePhaseRef.current >= 143) {
          isPulsingRef.current = false;
          pulsePhaseRef.current = 0;
          setIsPulsing(false);
        }
      }
    }, 35);

    // Draw function
    const draw = () => {
      // Semi-transparent black to create trail effect
      ctx.fillStyle = 'rgba(0, 0, 0, 0.05)';
      ctx.fillRect(0, 0, canvas.width, canvas.height);

      ctx.font = `${fontSize}px monospace`;

      // Draw characters
      for (let i = 0; i < activeColumns; i++) {
        // Random character
        const text = charArray[Math.floor(Math.random() * charArray.length)];
        
        // Calculate position
        const x = i * fontSize;
        const y = drops[i] * fontSize;

        // Check if this is the pulsing column and we're in pulse phase
        if (pulsingColumnRef.current === i && isPulsingRef.current) {
          // Pulse red for 5 seconds
          const pulseIntensity = Math.sin((pulsePhaseRef.current / 143) * Math.PI * 2) * 0.5 + 0.5;
          ctx.fillStyle = `rgba(255, 0, 0, ${0.5 + pulseIntensity * 0.5})`;
          ctx.shadowBlur = 15;
          ctx.shadowColor = '#ff0000';
          ctx.globalAlpha = 1.0;
        } else {
          // Draw character with varying opacity for depth
          const opacity = Math.random() * 0.5 + 0.5; // 0.5 to 1.0
          ctx.globalAlpha = opacity;
          ctx.fillStyle = '#00ff00';
          ctx.shadowBlur = 0;
        }

        ctx.fillText(text, x, y);

        // Reset drop to top with random delay
        if (y > canvas.height && Math.random() > 0.975) {
          drops[i] = 0;
        }

        // Move drop down
        drops[i]++;
      }

      ctx.globalAlpha = 1.0;
      ctx.shadowBlur = 0;
    };

    // Animation loop
    const interval = setInterval(draw, 35);

    return () => {
      clearInterval(interval);
      clearInterval(pulseInterval);
      clearInterval(pulseAnimation);
      window.removeEventListener('resize', resizeCanvas);
    };
  }, []); // Empty dependency array - effect only runs once

  const handleCanvasClick = (e: React.MouseEvent<HTMLCanvasElement>) => {
    if (!pulsingColumnRef.current || !isPulsingRef.current) return;

    const canvas = canvasRef.current;
    if (!canvas) return;

    const rect = canvas.getBoundingClientRect();
    const x = e.clientX - rect.left;

    const fontSize = 14;
    const clickColumn = Math.floor(x / fontSize);

    // Check if click is on the pulsing column
    if (clickColumn === pulsingColumnRef.current) {
      window.location.href = '/garbled';
    }
  };

  return (
    <canvas
      ref={canvasRef}
      className="fixed inset-0 w-full h-full z-0"
      style={{ background: '#000000', cursor: isPulsing && pulsingColumnRef.current !== null ? 'pointer' : 'default' }}
      onClick={handleCanvasClick}
    />
  );
};

