import {defineConfig, defaultColors} from 'histoire'
import {HstVue} from '@histoire/plugin-vue'
import {HstScreenshot} from '@histoire/plugin-screenshot'

export default defineConfig({
	setupFile: './src/histoire.setup.ts',
	storyIgnored: [
		'**/node_modules/**',
		'**/dist/**',
		// see https://kolaente.dev/vikunja/frontend/pulls/2724#issuecomment-42012
		'**/.direnv/**',
	],
	plugins: [
		HstVue(),
		HstScreenshot({
			// Options here
		}),
	],
	theme: {
		title: 'Vikunja',
		colors: {
			// https://histoire.dev/guide/config.html#builtin-colors
			gray: defaultColors.zinc,
			primary: defaultColors.cyan,
		},
		// logo: {
		// 	square: './img/square.png',
		// 	light: './img/light.png',
		// 	dark: './img/dark.png',
		// },
		logoHref: 'https://vikunja.io',
		// favicon: './favicon.ico',
	},
})
