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
	new RegExp('api\\/v1\\/.*$'),
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

const getBearerToken = async () => {
	// we can't get a client that sent the current request, therefore we need
	// to ask any controlled page for auth token
	const allClients = await self.clients.matchAll();
	const client = allClients.filter(client => client.type === 'window')[0];

	// if there is no page in scope, we can't get any token
	// and we indicate it with null value
	if(!client) {
		return null;
	}

	// to communicate with a page we will use MessageChannels
	// they expose pipe-like interface, where a receiver of
	// a message uses one end of a port for messaging and
	// we use the other end for listening
	const channel = new MessageChannel();

	client.postMessage({
		'action': 'getBearerToken'
	}, [channel.port1]);

	// ports support only onmessage callback which
	// is cumbersome to use, so we wrap it with Promise
	return new Promise((resolve, reject) => {
		channel.port2.onmessage = event => {
			if (event.data.error) {
				console.error('Port error', event.error);
				reject(event.data.error);
			}

			resolve(event.data.authToken);
		}
	});
}

// Notification action
self.addEventListener('notificationclick', function(event) {
	const taskId = event.notification.data.taskId
	event.notification.close()

	switch (event.action) {
		case 'mark-as-done':
			// FIXME: Ugly as hell, but no other way of doing this, since we can't use modules
			// in service workersfor now.
			fetch('/config.json')
				.then(r => r.json())
				.then(config => {

					getBearerToken()
						.then(token => {
							fetch(`${config.VIKUNJA_API_BASE_URL}tasks/${taskId}`, {
								method: 'post',
								headers: {
									'Accept': 'application/json',
									'Content-Type': 'application/json',
									'Authorization': `Bearer ${token}`,
								},
								body: JSON.stringify({id: taskId, done: true})
							})
							.then(r => r.json())
							.then(r => {
								console.debug('Task marked as done from notification', r)
							})
							.catch(e => {
								console.debug('Error marking task as done from notification', e)
							})
						})
				})
			break
		case 'show-task':
			clients.openWindow(`/tasks/${taskId}`)
			break
	}
})

workbox.core.clientsClaim();
// The precaching code provided by Workbox.
self.__precacheManifest = [].concat(self.__precacheManifest || []);
workbox.precaching.precacheAndRoute(self.__precacheManifest, {});
