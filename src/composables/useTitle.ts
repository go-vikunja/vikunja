import { computed, watchEffect } from 'vue'
import { setTitle } from '@/helpers/setTitle'

import { ComputedGetter, ComputedRef } from '@vue/reactivity'

export function useTitle<T>(titleGetter: ComputedGetter<T>) : ComputedRef<T> {
	const titleRef = computed(titleGetter)

	watchEffect(() => setTitle(titleRef.value))

	return titleRef
}