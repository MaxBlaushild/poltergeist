/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,jsx,ts,tsx}'],
  theme: {
    extend: {
      fontFamily: {
        // Decorative gothic display for hero titles, Cinzel for headings/labels,
        // and the very readable EB Garamond for body prose.
        display: ['"Cinzel Decorative"', 'Cinzel', 'serif'],
        heading: ['Cinzel', 'Georgia', 'serif'],
        serif: ['"EB Garamond"', 'Georgia', 'serif'],
        sans: ['"EB Garamond"', 'Georgia', 'serif'],
      },
      colors: {
        blood: {
          DEFAULT: '#8a0303',
          bright: '#ef5350', // brighter red — legible on black
          dim: '#5c0000',
          ink: '#0a0306',
        },
        // Warm parchment gold for small labels — much higher contrast than red.
        gold: '#e8c87a',
        bone: '#f2ebdc',
      },
    },
  },
  plugins: [],
};
