/** @type {import('tailwindcss').Config} */
export default {
  content: ["./index.html", "./src/**/*.{js,jsx}"],
  theme: {
    extend: {
      colors: {
        surface: "#2B2B2B",
        paper: "#EAEAEA",
        accent: "#FF3B30",
        line: "#565656",
        muted: "#A6A6A6",
      },
      fontFamily: {
        sans: ["Inter", "Helvetica Neue", "Helvetica", "Arial", "sans-serif"],
      },
      letterSpacing: {
        hero: "-0.06em",
      },
      minHeight: {
        hero: "calc(100vh - 70px)",
      },
    },
  },
  plugins: [],
};
