import {defineConfig, devices} from '@playwright/test'
import {execSync} from 'child_process'

// Find system chromium - for UI mode, set PLAYWRIGHT_CHROMIUM_EXECUTABLE_PATH env var
const getChromiumPath = () => {
	// Check if env var is already set (for UI mode)
	if (process.env.PLAYWRIGHT_CHROMIUM_EXECUTABLE_PATH) {
		return process.env.PLAYWRIGHT_CHROMIUM_EXECUTABLE_PATH
	}
	try {
		return execSync('which chromium', {encoding: 'utf-8'}).trim()
	} catch {
		return undefined
	}
}

export default defineConfig({
	testDir: './tests/e2e',
	fullyParallel: false,
	forbidOnly: !!process.env.CI,
	retries: process.env.CI ? 2 : 0,
	workers: 1, // No parallelization initially
	reporter: process.env.CI ? [['html'], ['list']] : 'html',
	use: {
		baseURL: 'http://127.0.0.1:4173',
		trace: 'on-first-retry',
		screenshot: 'only-on-failure',
		testIdAttribute: 'data-cy', // Preserve existing data-cy selectors
		launchOptions: {
			executablePath: getChromiumPath(),
		},
	},
	projects: [
		{
			name: 'chromium',
			use: {...devices['Desktop Chrome']},
		},
	],
	// webServer configuration removed - we manually start services in CI
	// For local development, run `pnpm preview` and `pnpm preview:vikunja` separately
})
