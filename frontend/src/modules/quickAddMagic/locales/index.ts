import {i18n, type SupportedLocale} from '@/i18n'

import {en} from './en'
import {uk} from './uk'
import type {QuickAddMagicLocale} from './types'

const LOCALES: Partial<Record<SupportedLocale, QuickAddMagicLocale>> = {
	'en': en,
	'uk-UA': uk,
}

/**
 * Resolves the parser locales for the current UI language. English is always
 * appended as a fallback because users commonly mix both languages.
 */
export function getDefaultLocales(): QuickAddMagicLocale[] {
	const primary = LOCALES[i18n.global.locale.value]
	return primary === undefined || primary.code === 'en' ? [en] : [primary, en]
}

export {en, uk}
export type {QuickAddMagicLocale}
