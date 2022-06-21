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
    baseUrl: 'http://localhost:4173',
    specPattern: 'cypress/e2e/**/*.{js,jsx,ts,tsx}',
  },
	component: {
    devServer: {
      framework: 'vue',
      bundler: 'vite',
    },
  },
})
