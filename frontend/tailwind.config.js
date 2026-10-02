/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        primary:   { DEFAULT: '#1976d2', dark: '#1565c0', light: '#42a5f5' },
        secondary: { DEFAULT: '#9c27b0', dark: '#7b1fa2' },
        success:   { DEFAULT: '#2e7d32', dark: '#1b5e20' },
        warn:      { DEFAULT: '#ed6c02', dark: '#e65100' },
        danger:    { DEFAULT: '#d32f2f', dark: '#c62828' },
      },
      fontFamily: {
        sans: ['Nunito', 'ui-rounded', 'SF Pro Rounded', 'system-ui', 'sans-serif'],
      },
      keyframes: {
        pop: { '0%': { transform: 'scale(0.9)' }, '60%': { transform: 'scale(1.05)' }, '100%': { transform: 'scale(1)' } },
        shake: { '0%,100%': { transform: 'translateX(0)' }, '25%': { transform: 'translateX(-4px)' }, '75%': { transform: 'translateX(4px)' } },
      },
      animation: {
        pop: 'pop 180ms ease-out',
        shake: 'shake 200ms ease-in-out',
      },
    },
  },
  plugins: [],
}
