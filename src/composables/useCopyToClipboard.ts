import {error} from '@/message'
import {useI18n} from 'vue-i18n'

export function useCopyToClipboard() {
	const {t} = useI18n()

	return async (text: string) => {
		try {
			await navigator.clipboard.writeText(text)
		} catch {
			error(t('misc.copyError'))
		}
	}
}