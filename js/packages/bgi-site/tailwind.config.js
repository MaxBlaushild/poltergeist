/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,jsx,ts,tsx}'],
  theme: {
    extend: {
      // A tabletop/cardboard palette — kraft paper, wood, felt playmat
      // green, dice-red — distinct from reef's aquarium teal/coral. Same
      // key names as reef's `colors.reef.*` on purpose: every reef
      // component's `bg-reef-*`/`text-reef-*` class translates 1:1 to
      // `bg-bgi-*`/`text-bgi-*` with no structural changes needed.
      colors: {
        bgi: {
          ink: '#241a12',
          deep: '#3a2a1c',
          lagoon: '#5c3d22',
          teal: '#7a5230',
          foam: '#f7f0e3',
          sand: '#e8dcc4',
          coral: '#b5432f',
          glow: '#8fae6b',
          sky: '#c9a876',
          paper: '#fffdf8',
        },
      },
      fontFamily: {
        display: ['"Fraunces"', 'Georgia', 'serif'],
      },
      boxShadow: {
        pop: '5px 5px 0 0 #241a12',
        'pop-sm': '3px 3px 0 0 #241a12',
        'pop-coral': '5px 5px 0 0 #b5432f',
        'pop-teal': '5px 5px 0 0 #7a5230',
        bgi: '0 20px 45px -15px rgba(58, 42, 28, 0.25)',
      },
      backgroundImage: {
        'bgi-hero':
          'radial-gradient(circle at 12% 15%, rgba(143,174,107,0.20), transparent 40%), radial-gradient(circle at 88% 10%, rgba(181,67,47,0.14), transparent 38%), radial-gradient(circle at 60% 90%, rgba(201,168,118,0.22), transparent 42%)',
      },
      keyframes: {
        popIn: {
          '0%': { transform: 'scale(0.9)', opacity: '0' },
          '60%': { transform: 'scale(1.03)', opacity: '1' },
          '100%': { transform: 'scale(1)', opacity: '1' },
        },
        shimmer: {
          '0%, 100%': { opacity: '0.6' },
          '50%': { opacity: '1' },
        },
      },
      animation: {
        shimmer: 'shimmer 3s ease-in-out infinite',
        'pop-in': 'popIn 0.5s cubic-bezier(0.34, 1.56, 0.64, 1) both',
      },
      transitionTimingFunction: {
        bounce: 'cubic-bezier(0.34, 1.56, 0.64, 1)',
      },
    },
  },
  plugins: [],
};
