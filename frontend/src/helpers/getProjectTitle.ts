import {i18n} from '@/i18n'
import type {Project} from '@/client/generated'

export function getProjectTitle(project: Required<Pick<Project, 'id' | 'title'>>) {
	if (project.id === -1) {
		return i18n.global.t('project.pseudo.favorites.title')
	}

	if (project.title === 'Inbox') {
		return i18n.global.t('project.inboxTitle')
	}

	if (project.title === 'My Open Tasks') {
		return i18n.global.t('project.myOpenTasksFilterTitle')
	}

	return project.title
}
