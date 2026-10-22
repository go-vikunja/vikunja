import {
	onScopeDispose,
	watch,
} from 'vue'
import {useWebSocket} from './useWebSocket'
import {useAuthStore} from '@/stores/auth'
import {
	CACHE_EVENTS,
	parseServerCacheEvent,
	useServerCacheEventMutation,
} from '@/client/queries/serverEvents'

export function useServerCacheEvents() {
	const socket = useWebSocket()
	const auth = useAuthStore()
	const mutation = useServerCacheEventMutation()
	const unsubscribers = CACHE_EVENTS.map(event =>
		socket.subscribe(event, message => {
			const update = parseServerCacheEvent(event, message.data, auth.info?.id)
			if (update) mutation.mutate(update)
		}),
	)
	let hasAuthenticated = socket.authenticated.value
	watch(socket.authenticated, authenticated => {
		if (!authenticated) return
		if (hasAuthenticated || socket.mayHaveMissedEvents.value) mutation.mutate({kind: 'reconnect'})
		else mutation.mutate({
			kind: 'subscribed',
			since: socket.subscribedAt.value,
		})
		hasAuthenticated = true
	})
	onScopeDispose(() => unsubscribers.forEach(unsubscribe => unsubscribe()))
}
