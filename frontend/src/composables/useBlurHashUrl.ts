import {shallowRef, toValue, watch, type MaybeRefOrGetter} from 'vue'
import {useObjectUrl} from '@vueuse/core'

import {getBlobFromBlurHash} from '@/helpers/getBlobFromBlurHash'

export function useBlurHashUrl(hash: MaybeRefOrGetter<string>) {
	const blob = shallowRef<Blob | null>(null)

	watch(() => toValue(hash), async (value, _previous, onCleanup) => {
		let current = true
		onCleanup(() => {
			current = false
		})

		blob.value = null
		if (value === '') {
			return
		}

		try {
			const decoded = await getBlobFromBlurHash(value)
			if (current) {
				blob.value = decoded
			}
		} catch (e) {
			console.error('Error generating blur hash preview', e)
		}
	}, {immediate: true})

	return useObjectUrl(blob)
}
