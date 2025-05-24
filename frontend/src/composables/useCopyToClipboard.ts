import {error} from '@/message'
import {useI18n} from 'vue-i18n'

export function useCopyToClipboard() {
	const {t} = useI18n({useScope: 'global'})
	
	function fallbackCopyTextToClipboard(text: string) {
		const textArea = document.createElement('textarea')
		textArea.value = text
		
		// Avoid scrolling to bottom
		textArea.style.top = '0'
		textArea.style.left = '0'
		textArea.style.position = 'fixed'
	
		document.body.appendChild(textArea)
		textArea.focus()
		textArea.select()
	
		try {
			// NOTE: the execCommand is deprecated but as of 2022_09
			// widely supported and works without https
			const successful = document.execCommand('copy')
			if (!successful) {
				throw new Error()
			}
			// eslint-disable-next-line @typescript-eslint/no-unused-vars
		} catch (e) {
			error(t('misc.copyError'))
		}
	
		document.body.removeChild(textArea)
	}
	
	return async (text: string) => {
		if (!navigator.clipboard) {
			fallbackCopyTextToClipboard(text)
			return
		}
		try {
			await navigator.clipboard.writeText(text)
			// eslint-disable-next-line @typescript-eslint/no-unused-vars
		} catch(e) {
			error(t('misc.copyError'))
		}
	}
}
