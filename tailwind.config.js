/** @type {import('tailwindcss').Config} */
module.exports = {
  darkMode: 'class',
  content: ["./site/**/*.{html,js}"],
  theme: {
    extend: {
      animation: {
        fade: 'fadeOut 5s ease-in-out',
      },

      keyframes: theme => ({
        fadeOut: {
          '0%': { opacity: 1 },
          '100%': { opacity: 0 },
        },
      }),
    },
    fontFamily: {
      'roboto': ['Roboto', 'sans-serif'],
    }
  },
  plugins: [],
}

