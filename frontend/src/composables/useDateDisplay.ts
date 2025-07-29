import {computed} from 'vue'
import {createSharedComposable} from '@vueuse/core'
import {useAuthStore} from '@/stores/auth'

export const useDateDisplay = createSharedComposable(() => {
	const authStore = useAuthStore()
	const store = computed(() => authStore.settings.frontendSettings.dateDisplay)
	return {store}
})
