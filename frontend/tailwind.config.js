/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        // Brand Cinema Accents
        oryn: {
          50: '#FDF8F0',
          100: '#F9ECD9',
          200: '#F2D7B0',
          300: '#EABF83',
          400: '#E2A456',
          500: '#D98226',
          600: '#B86518',
          700: '#944C15',
          800: '#783D18',
          900: '#643417',
          gold: '#E5A65E',
          accent: '#F59E0B',
        },
        // Cinema Dark Surfaces
        cinema: {
          950: '#07090E',
          900: '#0B0F17',
          850: '#111722',
          800: '#172030',
          750: '#1E293D',
          700: '#26334D',
          border: '#1F2A3D',
          hover: '#293750',
        },
        // Clean Light Surfaces
        light: {
          bg: '#F8FAFC',
          card: '#FFFFFF',
          cardHover: '#F1F5F9',
          border: '#E2E8F0',
          textPrimary: '#0F172A',
          textSecondary: '#475569',
          textMuted: '#94A3B8',
        }
      },
      fontFamily: {
        sans: ['Vazirmatn', 'Inter', '-apple-system', 'BlinkMacSystemFont', 'Segoe UI', 'Roboto', 'sans-serif'],
        fa: ['Vazirmatn', 'sans-serif'],
      },
      boxShadow: {
        cinema: '0 8px 30px rgba(0, 0, 0, 0.45)',
        goldGlow: '0 0 20px rgba(229, 166, 94, 0.25)',
      }
    },
  },
  plugins: [],
}
