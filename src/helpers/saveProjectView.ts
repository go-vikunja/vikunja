import type { IProject } from '@/modelTypes/IProject'

type ProjectView = Record<IProject['id'], string>

const DEFAULT_PROJECT_VIEW = 'project.list' as const
const PROJECT_VIEW_SETTINGS_KEY = 'projectView'

/**
 * Save the current project view to local storage
 */
export function saveProjectView(projectId: IProject['id'], routeName: string) {
	if (routeName.includes('settings.')) {
		return
	}
	
	if (!projectId) {
		return
	}
	
	// We use local storage and not the store here to make it persistent across reloads.
	const savedProjectView = localStorage.getItem(PROJECT_VIEW_SETTINGS_KEY)
	let savedProjectViewJson: ProjectView | false = false
	if (savedProjectView !== null) {
		savedProjectViewJson = JSON.parse(savedProjectView) as ProjectView
	}

	let projectView: ProjectView = {}
	if (savedProjectViewJson) {
		projectView = savedProjectViewJson
	}

	projectView[projectId] = routeName
	localStorage.setItem(PROJECT_VIEW_SETTINGS_KEY, JSON.stringify(projectView))
}

export const getProjectView = (projectId: IProject['id']) => {
	// Migrate old setting over
	// TODO: remove when 1.0 release
	const oldListViewSettings = localStorage.getItem('listView')
	if (oldListViewSettings !== null) {
		localStorage.setItem(PROJECT_VIEW_SETTINGS_KEY, oldListViewSettings)
		localStorage.removeItem('listView')
	}
	
	// Remove old stored settings
	// TODO: remove when 1.0 release
	const savedProjectView = localStorage.getItem(PROJECT_VIEW_SETTINGS_KEY)
	if (savedProjectView !== null && savedProjectView.startsWith('project.')) {
		localStorage.removeItem(PROJECT_VIEW_SETTINGS_KEY)
	}

	if (!savedProjectView) {
		return DEFAULT_PROJECT_VIEW
	}

	const savedProjectViewJson: ProjectView = JSON.parse(savedProjectView)

	if (!savedProjectViewJson[projectId]) {
		return DEFAULT_PROJECT_VIEW
	}

	return savedProjectViewJson[projectId]
}