import {onScopeDispose, readonly, ref} from 'vue'

// useMounted() never resets on unmount.
export function useIsAlive() {
	const alive = ref(true)
	onScopeDispose(() => {
		alive.value = false
	})
	return readonly(alive)
}
