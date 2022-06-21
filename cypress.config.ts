import { defineConfig } from 'cypress'

export default defineConfig({
  env: {
    API_URL: 'http://localhost:3456/api/v1',
    TEST_SECRET: 'averyLongSecretToSe33dtheDB',
  },
  video: false,
  retries: {
    runMode: 2,
  },
  projectId: '181c7x',
  e2e: {
    // We've imported your old cypress plugins here.
    // You may want to clean this up later by importing these.
    setupNodeEvents(on, config) {
      return require('./cypress/plugins/index.ts')(on, config)
    },
    baseUrl: 'http://localhost:4173',
    specPattern: 'cypress/e2e/**/*.{js,jsx,ts,tsx}',
  },
})
