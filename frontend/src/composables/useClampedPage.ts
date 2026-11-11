import {
	watch,
	type Ref,
} from 'vue'
import {
	clampPage,
	type Paginated,
} from '@/client/queries/pagination'

export function useClampedPage(page: Ref<number>, query: {
	data: Readonly<Ref<Pick<Paginated<unknown>, 'total_pages'> | undefined>>,
	isPlaceholderData: Readonly<Ref<boolean>>,
}) {
	watch([query.data, query.isPlaceholderData], ([data, isPlaceholderData]) => {
		const clamped = clampPage(page.value, {
			data,
			isPlaceholderData,
		})
		if (clamped !== page.value) page.value = clamped
	}, {immediate: true})
}
