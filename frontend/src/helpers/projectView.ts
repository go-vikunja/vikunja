import router from '@/router'

import type {IProject} from '@/modelTypes/IProject'

export type ProjectViewSettings = Record<IProject['id'], number>

const SETTINGS_KEY_PROJECT_VIEW = 'projectView'

/**
 * Save the current project view to local storage
 */
export function saveProjectView(projectId: IProject['id'], viewId: number) {
	if (!projectId || !viewId) {
		return
	}
	
	// We use local storage and not the store here to make it persistent across reloads.
	const savedProjectView = localStorage.getItem(SETTINGS_KEY_PROJECT_VIEW)
	let savedProjectViewSettings: ProjectViewSettings | false = false
	if (savedProjectView !== null) {
		savedProjectViewSettings = JSON.parse(savedProjectView) as ProjectViewSettings
	}

	let projectViewSettings: ProjectViewSettings = {}
	if (savedProjectViewSettings) {
		projectViewSettings = savedProjectViewSettings
	}

	projectViewSettings[projectId] = viewId
	localStorage.setItem(SETTINGS_KEY_PROJECT_VIEW, JSON.stringify(projectViewSettings))
}

export const getProjectView = (projectId: IProject['id']) => {
	try {	
		const projectViewSettingsString = localStorage.getItem(SETTINGS_KEY_PROJECT_VIEW)
		if (!projectViewSettingsString) {
			throw new Error()
		}
		
		const projectViewSettings = JSON.parse(projectViewSettingsString) as ProjectViewSettings
		if (!router.hasRoute(projectViewSettings[projectId])) {
			throw new Error()
		}
		return projectViewSettings[projectId]
	} catch (e) {
		return
	}	
}