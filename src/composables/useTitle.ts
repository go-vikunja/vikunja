import { computed, watchEffect } from 'vue'
import type { ComputedGetter } from 'vue'

import { setTitle } from '@/helpers/setTitle'

export function useTitle(titleGetter: ComputedGetter<string>) {
	const titleRef = computed(titleGetter)

	watchEffect(() => setTitle(titleRef.value))

	return titleRef
}