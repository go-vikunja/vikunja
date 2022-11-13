import {i18n} from '@/i18n'
import type {INamespace} from '@/modelTypes/INamespace'

export const getNamespaceTitle = (n: INamespace) => {
	if (n.id === -1) {
		return i18n.global.t('namespace.pseudo.sharedProjects.title')
	}
	if (n.id === -2) {
		return i18n.global.t('namespace.pseudo.favorites.title')
	}
	if (n.id === -3) {
		return i18n.global.t('namespace.pseudo.savedFilters.title')
	}
	return n.title
}
