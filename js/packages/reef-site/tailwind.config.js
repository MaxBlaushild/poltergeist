/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,jsx,ts,tsx}'],
  theme: {
    extend: {
      colors: {
        reef: {
          ink: '#0b1c26',
          deep: '#0f2a38',
          teal: '#1f6f78',
          coral: '#ff7a59',
          sand: '#f4ede1',
        },
      },
    },
  },
  plugins: [],
};
