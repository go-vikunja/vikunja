import { computed, watchEffect } from 'vue'
import { setTitle } from '@/helpers/setTitle'

import { ComputedGetter } from '@vue/reactivity'

export function useTitle(titleGetter: ComputedGetter<string>) {
	const titleRef = computed(titleGetter)

	watchEffect(() => setTitle(titleRef.value))

	return titleRef
}