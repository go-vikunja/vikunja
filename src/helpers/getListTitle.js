import {i18n} from '@/i18n'

export const getListTitle = (l) => {
	if (l.id === -1) {
		return i18n.global.t('list.pseudo.favorites.title')
	}
	return l.title
}
