import {computed} from 'vue'
import {useI18n} from 'vue-i18n'
import {isRTLLanguage, type SupportedLocale} from '@/i18n'

export function useIsRtl() {
	const {locale} = useI18n()
	return computed(() => isRTLLanguage(locale.value as SupportedLocale))
}
