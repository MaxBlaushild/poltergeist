/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,jsx,ts,tsx}'],
  theme: {
    extend: {
      colors: {
        reef: {
          ink: '#0b1c26',
          deep: '#0f2a38',
          lagoon: '#0e4a56',
          teal: '#1f6f78',
          foam: '#eefaf7',
          sand: '#f4ede1',
          coral: '#ff7a59',
          glow: '#57e6c9',
          sky: '#5aa9ff',
          paper: '#ffffff',
        },
      },
      fontFamily: {
        display: ['"Space Grotesk"', 'system-ui', 'sans-serif'],
      },
      boxShadow: {
        // Hard, offset "sticker" shadows — no blur — for the pressable,
        // bouncy button/card feel, in place of the old soft glow shadows.
        pop: '5px 5px 0 0 #0b1c26',
        'pop-sm': '3px 3px 0 0 #0b1c26',
        'pop-coral': '5px 5px 0 0 #ff7a59',
        'pop-teal': '5px 5px 0 0 #1f6f78',
        reef: '0 20px 45px -15px rgba(15, 42, 56, 0.2)',
      },
      backgroundImage: {
        'reef-hero': 'radial-gradient(circle at 12% 15%, rgba(87,230,201,0.25), transparent 40%), radial-gradient(circle at 88% 10%, rgba(255,122,89,0.18), transparent 38%), radial-gradient(circle at 60% 90%, rgba(90,169,255,0.16), transparent 42%)',
      },
      keyframes: {
        drift: {
          '0%, 100%': { transform: 'translateY(0) translateX(0) rotate(0deg)' },
          '50%': { transform: 'translateY(-18px) translateX(8px) rotate(4deg)' },
        },
        blob: {
          '0%, 100%': { borderRadius: '42% 58% 65% 35% / 45% 40% 60% 55%', transform: 'scale(1)' },
          '50%': { borderRadius: '58% 42% 35% 65% / 55% 65% 35% 45%', transform: 'scale(1.05)' },
        },
        rise: {
          '0%': { transform: 'translateY(0)', opacity: '0.7' },
          '100%': { transform: 'translateY(-140px)', opacity: '0' },
        },
        shimmer: {
          '0%, 100%': { opacity: '0.6' },
          '50%': { opacity: '1' },
        },
        popIn: {
          '0%': { transform: 'scale(0.9)', opacity: '0' },
          '60%': { transform: 'scale(1.03)', opacity: '1' },
          '100%': { transform: 'scale(1)', opacity: '1' },
        },
      },
      animation: {
        drift: 'drift 7s ease-in-out infinite',
        'drift-slow': 'drift 10s ease-in-out infinite',
        blob: 'blob 8s ease-in-out infinite',
        'blob-slow': 'blob 12s ease-in-out infinite',
        rise: 'rise 7s linear infinite',
        shimmer: 'shimmer 3s ease-in-out infinite',
        'pop-in': 'popIn 0.5s cubic-bezier(0.34, 1.56, 0.64, 1) both',
      },
      transitionTimingFunction: {
        // A "back-out" spring — overshoots slightly past its target before
        // settling, which is what actually reads as "bouncy" rather than
        // just fast.
        bounce: 'cubic-bezier(0.34, 1.56, 0.64, 1)',
      },
    },
  },
  plugins: [],
};
