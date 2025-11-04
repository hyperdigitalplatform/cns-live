/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  darkMode: 'selector',
  theme: {
    extend: {
      colors: {
        primary: {
          50: '#f0f9ff',
          100: '#e0f2fe',
          200: '#bae6fd',
          300: '#7dd3fc',
          400: '#38bdf8',
          500: '#0ea5e9',
          600: '#0284c7',
          700: '#0369a1',
          800: '#075985',
          900: '#0c4a6e',
        },
        dark: {
          base: '#1C1C1E',        // Darker base for better depth
          secondary: '#2C2C2E',    // Keep
          sidebar: '#252527',      // Slightly lighter sidebar
          surface: '#3A3A3C',      // Medium surface
          border: '#48484D',       // Keep
          elevated: '#52525A',     // Noticeably lighter for hover (was #52525B)
          hover: '#5A5A62',        // NEW: Even more obvious hover state
        },
        text: {
          primary: '#FFFFFF',
          secondary: '#A1A1AA',
          muted: '#71717A',
        },
      },
    },
  },
  plugins: [],
}
