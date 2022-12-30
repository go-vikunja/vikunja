/// <reference types="vitest" />
import {defineConfig, type PluginOption} from 'vite'
import vue from '@vitejs/plugin-vue'
import legacyFn from '@vitejs/plugin-legacy'
import { URL, fileURLToPath } from 'node:url'
import { dirname, resolve } from 'node:path'

import VueI18nPlugin from '@intlify/unplugin-vue-i18n/vite'
import {VitePWA}  from 'vite-plugin-pwa'
import {visualizer}  from 'rollup-plugin-visualizer'
import svgLoader from 'vite-svg-loader'
import postcssPresetEnv from 'postcss-preset-env'


const pathSrc = fileURLToPath(new URL('./src', import.meta.url))

// the @use rules have to be the first in the compiled stylesheets
const PREFIXED_SCSS_STYLES = `@use "sass:math";
@import "${pathSrc}/styles/common-imports";`

const isModernBuild = Boolean(process.env.BUILD_MODERN_ONLY)
const legacy = isModernBuild
	? undefined
	: legacyFn({
		// recommended by browserslist => https://github.com/vitejs/vite/tree/main/packages/plugin-legacy#targets
		targets: ['defaults', 'not IE 11'],
	})

console.log(isModernBuild
	? 'Building "modern-only" build'
	: 'Building "legacy" build with "@vitejs/plugin-legacy"',
)

// https://vitejs.dev/config/
export default defineConfig({
	// https://vitest.dev/config/
	test: {
		environment: 'happy-dom',
	},
	css: {
		preprocessorOptions: {
			scss: {
				additionalData: PREFIXED_SCSS_STYLES,
				charset: false, // fixes  "@charset" must be the first rule in the file" warnings
			},
		},
		postcss: {
			plugins: [
				postcssPresetEnv(),
			],
		},
	},
	plugins: [
		vue({
			reactivityTransform: true,
		}),
		legacy,
		svgLoader({
			// Since the svgs are already manually optimized via https://jakearchibald.github.io/svgomg/
			// we don't need to optimize them again.
			svgo: false,
		}),
		VueI18nPlugin({
			// TODO: only install needed stuff
			// Whether to install the full set of APIs, components, etc. provided by Vue I18n.
			// By default, all of them will be installed.
			fullInstall: true,
			include: resolve(dirname(pathSrc), './src/i18n/lang/**'),
		}),
		VitePWA({
			srcDir: 'src',
			filename: 'sw.ts',
			base: '/',
			strategies: 'injectManifest',
			injectRegister: false,
			manifest: {
				name: 'Vikunja',
				short_name: 'Vikunja',
				theme_color: '#1973ff',
				icons: [
					{
						src: './images/icons/android-chrome-192x192.png',
						sizes: '192x192',
						type: 'image/png',
					},
					{
						src: './images/icons/android-chrome-512x512.png',
						sizes: '512x512',
						type: 'image/png',
					},
					{
						src: './images/icons/icon-maskable.png',
						sizes: '1024x1024',
						type: 'image/png',
						purpose: 'maskable',
					},
				],
				start_url: '.',
				display: 'standalone',
				background_color: '#000000',
				shortcuts: [
					{
						name: 'Overview',
						url: '/',
					},
					{
						name: 'Namespaces And Lists Overview',
						short_name: 'Namespaces & Lists',
						url: '/namespaces',
					},
					{
						name: 'Tasks Next Week',
						short_name: 'Next Week',
						url: '/tasks/by/week',
					},
					{
						name: 'Tasks Next Month',
						short_name: 'Next Month',
						url: '/tasks/by/month',
					},
					{
						name: 'Teams Overview',
						short_name: 'Teams',
						url: '/teams',
					},
				],
			},
		}),
	],
	resolve: {
		alias: [
			{
				find: '@',
				replacement: pathSrc,
			},
		],
		extensions: ['.mjs', '.js', '.ts', '.jsx', '.tsx', '.json', '.vue'],
	},
	server: {
		port: 4173,
		strictPort: true,
	},
	build: {
		target: 'esnext',
		rollupOptions: {
			plugins: [
				visualizer({
					filename: 'stats.html',
					gzipSize: true,
					// template: 'sunburst',
					// brotliSize: true,
				}) as PluginOption,
			],
		},
	},
})
