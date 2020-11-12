module.exports = {
	configureWebpack: {
		devtool: 'source-map',
	},
	productionSourceMap: false,
	pwa: {
		name: 'Vikunja',
		themeColor: '#1973ff',
		appleMobileWebAppCapable: 'yes',
		workboxPluginMode: 'InjectManifest',
		workboxOptions: {
			importWorkboxFrom: 'local',
			swSrc: 'src/ServiceWorker/sw.js',
		},
		iconPaths: {
			favicon32: 'images/icons/favicon-32x32.png',
			favicon16: 'images/icons/favicon-16x16.png',
			appleTouchIcon: 'images/icons/apple-touch-icon-152x152.png',
			maskIcon: 'images/icons/safari-pinned-tab.svg',
			msTileImage: 'images/icons/msapplication-icon-144x144.png',
		},
		manifestOptions: {
			"icons": [
				{
					"src": "./images/icons/android-chrome-192x192.png",
					"sizes": "192x192",
					"type": "image/png"
				},
				{
					"src": "./images/icons/android-chrome-512x512.png",
					"sizes": "512x512",
					"type": "image/png"
				},
				{
					"src": "./images/icons/icon-maskable.png",
					"sizes": "1024x1024",
					"type": "image/png",
					"purpose": "maskable"
				},
			],
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
				}
			]
		},
	}
}