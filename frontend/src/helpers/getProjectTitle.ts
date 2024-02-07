import {i18n} from '@/i18n'
import type {IProject} from '@/modelTypes/IProject'

export function getProjectTitle(project: IProject) {
	if (project.id === -1) {
		return i18n.global.t('project.pseudo.favorites.title')
	}

	if (project.title === 'Inbox') {
		return i18n.global.t('project.inboxTitle')
	}

	return project.title
}
