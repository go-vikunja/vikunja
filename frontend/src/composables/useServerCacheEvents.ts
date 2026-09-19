import {
	onScopeDispose,
	watch,
} from 'vue'
import {useWebSocket} from './useWebSocket'
import {useAuthStore} from '@/stores/auth'
import {
	parseServerCacheEvent,
	useServerCacheEventMutation,
} from '@/client/queries/serverEvents'

export function useServerCacheEvents() {
	const socket = useWebSocket()
	const auth = useAuthStore()
	const mutation = useServerCacheEventMutation()
	const unsubscribers = ['timer.created', 'timer.updated', 'timer.deleted', 'notification.created'].map(event =>
		socket.subscribe(event, message => {
			const update = parseServerCacheEvent(event, message.data, auth.info?.id)
			if (update) mutation.mutate(update)
		}),
	)
	// Only a re-authentication after a drop can have missed events; the first one follows the initial fetches.
	let hasAuthenticated = socket.authenticated.value
	watch(socket.authenticated, authenticated => {
		if (!authenticated) return
		if (hasAuthenticated) mutation.mutate({kind: 'reconnect'})
		hasAuthenticated = true
	})
	onScopeDispose(() => unsubscribers.forEach(unsubscribe => unsubscribe()))
}
