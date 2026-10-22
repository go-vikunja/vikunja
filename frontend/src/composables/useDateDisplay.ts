import {computed} from 'vue'
import {createSharedComposable} from '@vueuse/core'
import {useAuthStore} from '@/stores/auth'

export const useDateDisplay = createSharedComposable(() => {
	const authStore = useAuthStore()
	const store = computed(() => authStore.settings.frontend_settings.date_display)
	return {store}
})
