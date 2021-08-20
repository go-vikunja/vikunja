import createVuePlugin from '@vitejs/plugin-vue'
const {VitePWA} = require('vite-plugin-pwa')
const path = require('path')
const {visualizer} = require('rollup-plugin-visualizer')

const pathSrc = path.resolve(__dirname, './src')

// the @use rules have to be the first in the compiled stylesheets
const SCSS_IMPORT_PREFIX = `@use "sass:math";
@import "${pathSrc}/styles/variables";`

module.exports = {
	css: {
		preprocessorOptions: {
			scss: { additionalData: SCSS_IMPORT_PREFIX },
		},
	},
	plugins: [
		createVuePlugin({
			template: {
				compilerOptions: {
					compatConfig: {
						MODE: 2,
					},
				},
			},
		}),
		VitePWA({
			srcDir: 'src',
			filename: 'sw.js',
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
				find: 'vue',
				replacement: '@vue/compat',
			},
			{
				find: '@',
				replacement: path.resolve(__dirname, 'src'),
			},
		],
		extensions: ['.mjs', '.js', '.ts', '.jsx', '.tsx', '.json', '.vue'],
	},
	server: {
		port: 5000,
		strictPort: true,
	},
	build: {
		target: 'es2015',
		rollupOptions: {
			plugins: [
				visualizer({
					filename: 'stats.html',
				}),
			],
		},
	},
}
