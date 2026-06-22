/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
        ink: {
          950: '#05080d',
          900: '#090e16',
          850: '#0d141f',
          800: '#111b29',
          700: '#1a293b',
        },
        signal: {
          DEFAULT: '#9ef01a',
          dim: '#6ca50e',
        },
        cyan: {
          350: '#4de7e2',
        },
      },
      boxShadow: {
        glow: '0 0 32px rgba(158, 240, 26, 0.14)',
        panel: '0 20px 60px rgba(0, 0, 0, 0.28)',
      },
      fontFamily: {
        sans: ['Inter', 'ui-sans-serif', 'system-ui', 'sans-serif'],
        mono: ['JetBrains Mono', 'SFMono-Regular', 'Consolas', 'monospace'],
      },
    },
  },
  plugins: [],
}

