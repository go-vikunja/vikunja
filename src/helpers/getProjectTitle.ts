import {i18n} from '@/i18n'
import type {IProject} from '@/modelTypes/IProject'

export function getProjectTitle(p: IProject) {
	if (p.id === -1) {
		return i18n.global.t('project.pseudo.favorites.title')
	}
	
	if (p.title === 'Inbox') {
		return i18n.global.t('project.inboxTitle')
	}
	
	return p.title
}
