import {i18n} from '@/i18n'

export const getNamespaceTitle = (n) => {
	if (n.id === -1) {
		return i18n.global.t('namespace.pseudo.sharedLists.title')
	}
	if (n.id === -2) {
		return i18n.global.t('namespace.pseudo.favorites.title')
	}
	if (n.id === -3) {
		return i18n.global.t('namespace.pseudo.savedFilters.title')
	}
	return n.title
}
