import {onScopeDispose, watch} from 'vue'
import {useWebSocket} from './useWebSocket'
import {useAuthStore} from '@/stores/auth'
import {parseServerCacheEvent, useServerCacheEventMutation} from '@/client/queries/serverEvents'

export function useServerCacheEvents() {
	const socket = useWebSocket()
	const auth = useAuthStore()
	const mutation = useServerCacheEventMutation()
	const unsubscribers = ['timer.created', 'timer.updated', 'timer.deleted', 'notification.created'].map(event =>
		socket.subscribe(event, message => {
			const update = parseServerCacheEvent(event, message.data)
			if (!update || ('entry' in update && update.entry.user_id !== auth.info?.id)) return
			mutation.mutate(update)
		}),
	)
	watch(socket.authenticated, authenticated => {
		if (authenticated) mutation.mutate({kind: 'reconnect'})
	}, {immediate: true})
	onScopeDispose(() => unsubscribers.forEach(unsubscribe => unsubscribe()))
}
