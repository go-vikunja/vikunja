import {useAuthStore} from '@/stores/auth'
import {useBaseStore} from '@/stores/base'

export async function loadUserInfo() {
	// appReady doesn't wait on router.isReady(), so this can't deadlock the initial navigation.
	await useBaseStore().appReady
	const authStore = useAuthStore()
	if (!authStore.currentAccount) {
		await authStore.refreshUserInfo()
	}
	return authStore.currentAccount
}
