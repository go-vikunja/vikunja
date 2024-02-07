import {computed} from 'vue'
import type {Ref} from 'vue'

import {useTitle as useTitleVueUse, toRef} from '@vueuse/core'

type UseTitleParameters = Parameters<typeof useTitleVueUse>

export function useTitle(...args: UseTitleParameters) {

	const [newTitle, ...restArgs] = args

	const pageTitle = toRef(newTitle) as Ref<string>

	const completeTitle = computed(() =>
		(typeof pageTitle.value === 'undefined' || pageTitle.value === '')
			? 'Vikunja'
			: `${pageTitle.value} | Vikunja`,
	)

	return useTitleVueUse(completeTitle, ...restArgs)
}