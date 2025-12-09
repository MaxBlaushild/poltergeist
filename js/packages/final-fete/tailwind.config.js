/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,jsx,ts,tsx}'],
  theme: {
    extend: {
      fontFamily: {
        mono: ['Courier New', 'Monaco', 'Consolas', 'Lucida Console', 'monospace'],
        sans: ['Courier New', 'Monaco', 'Consolas', 'Lucida Console', 'monospace'],
      },
      colors: {
        matrix: {
          green: '#00ff00',
          'green-bright': '#00ff41',
          'green-dim': '#00cc00',
          black: '#000000',
        },
      },
    },
  },
  plugins: [],
};

