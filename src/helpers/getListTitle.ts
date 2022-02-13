import {i18n} from '@/i18n'

import ListModal from '@/modals/list'

export function getListTitle(l: ListModal) {
	if (l.id === -1) {
		return i18n.global.t('list.pseudo.favorites.title')
	}
	return l.title
}
