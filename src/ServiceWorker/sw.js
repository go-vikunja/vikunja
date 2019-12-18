/* eslint-disable no-console */
/* eslint-disable no-undef */

// Cache assets
workbox.routing.registerRoute(
    // This regexp matches all files in precache-manifest
    new RegExp('.+\\.(css|json|js|eot|svg|ttf|woff|woff2|png|html|txt)$'),
    new workbox.strategies.StaleWhileRevalidate()
);

// Always send api reqeusts through the network
workbox.routing.registerRoute(
	new RegExp('(\\/)?api\\/v1\\/.*$'),
	new workbox.strategies.NetworkOnly()
);

// Cache everything else
workbox.routing.registerRoute(
	new RegExp('.*'),
    new workbox.strategies.StaleWhileRevalidate()
);

// This code listens for the user's confirmation to update the app.
self.addEventListener('message', (e) => {
	if (!e.data) {
		return;
	}

	switch (e.data) {
		case 'skipWaiting':
			self.skipWaiting();
			break;
		default:
			// NOOP
			break;
	}
});

workbox.core.clientsClaim();
// The precaching code provided by Workbox.
self.__precacheManifest = [].concat(self.__precacheManifest || []);
workbox.precaching.precacheAndRoute(self.__precacheManifest, {});
