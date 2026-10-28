import {computed} from 'vue'
import {useI18n} from 'vue-i18n'

import {useConfigStore} from '@/stores/config'

const TRANSLATED_TEXTS = new Map([
	['Imprint', 'navigation.imprint'],
	['Privacy Policy', 'navigation.privacy'],
])

export function useLegalLinks() {
	const configStore = useConfigStore()
	const {t} = useI18n({useScope: 'global'})

	return computed(() => configStore.legal.links.map(link => {
		const key = TRANSLATED_TEXTS.get(link.text)
		return {
			url: link.url,
			text: key ? t(key) : link.text,
		}
	}))
}
