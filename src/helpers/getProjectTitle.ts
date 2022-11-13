import {i18n} from '@/i18n'
import type {IProject} from '@/modelTypes/IProject'

export function getProjectTitle(l: IProject) {
	if (l.id === -1) {
		return i18n.global.t('project.pseudo.favorites.title')
	}
	return l.title
}
