import {computed, type ComputedRef} from 'vue'
import {useConfigStore} from '@/stores/config'

/**
 * Returns a reactive boolean indicating whether a licensed pro feature is enabled.
 * The list comes from the /info endpoint and is populated into the config store on startup.
 */
export function useFeature(name: string): ComputedRef<boolean> {
	const store = useConfigStore()
	return computed(() => store.enabledProFeatures?.includes(name) ?? false)
}
