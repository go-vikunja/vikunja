/** @type {import('tailwindcss').Config} */
module.exports = {
  corePlugins: {
    // TODO: Re-add after removing bulma base styles
    preflight: false,
  },
  prefix: 'tw-',
  content: [
    './index.html',
    './src/**/*.{vue,js,ts}',
  ],
  theme: {
    extend: {},
  },
  plugins: [],
}

