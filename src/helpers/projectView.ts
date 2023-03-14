import type { RouteRecordName } from 'vue-router'
import router from '@/router'

import type {IProject} from '@/modelTypes/IProject'

export type ProjectRouteName = Extract<RouteRecordName, string>
export type ProjectViewSettings = Record<
	IProject['id'],
	Extract<RouteRecordName, ProjectRouteName>
>

const SETTINGS_KEY_PROJECT_VIEW = 'projectView'

// TODO: remove migration when releasing 1.0 
type ListViewSettings = ProjectViewSettings
const SETTINGS_KEY_DEPRECATED_LIST_VIEW = 'listView'
function migrateStoredProjectRouteSettings() {
	try {
		const listViewSettingsString = localStorage.getItem(SETTINGS_KEY_DEPRECATED_LIST_VIEW)
		if (listViewSettingsString === null) {
			return 
		}

		// A) the first version stored one setting for all lists in a string
		if (listViewSettingsString.startsWith('list.')) {
			const projectView = listViewSettingsString.replace('list.', 'project.')

			if (!router.hasRoute(projectView)) {
				return
			}
			return projectView as RouteRecordName
		}

		// B) the last version used a 'list.' prefix
		const listViewSettings: ListViewSettings = JSON.parse(listViewSettingsString)

		const projectViewSettingEntries = Object.entries(listViewSettings).map(([id, value]) => {
			return [id, value.replace('list.', 'project.')]
		})
		const projectViewSettings = Object.fromEntries(projectViewSettingEntries)

		localStorage.setItem(SETTINGS_KEY_PROJECT_VIEW, JSON.stringify(projectViewSettings))
	} catch(e) {
		// 
	} finally {
		localStorage.removeItem(SETTINGS_KEY_DEPRECATED_LIST_VIEW)
	}
}

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
	const savedProjectView = localStorage.getItem(SETTINGS_KEY_PROJECT_VIEW)
	let savedProjectViewSettings: ProjectViewSettings | false = false
	if (savedProjectView !== null) {
		savedProjectViewSettings = JSON.parse(savedProjectView) as ProjectViewSettings
	}

	let projectViewSettings: ProjectViewSettings = {}
	if (savedProjectViewSettings) {
		projectViewSettings = savedProjectViewSettings
	}

	projectViewSettings[projectId] = routeName
	localStorage.setItem(SETTINGS_KEY_PROJECT_VIEW, JSON.stringify(projectViewSettings))
}

export const getProjectView = (projectId: IProject['id']) => {
	// TODO: remove migration when releasing 1.0 
	const migratedProjectView = migrateStoredProjectRouteSettings()

	if (migratedProjectView !== undefined && router.hasRoute(migratedProjectView))  {
		return migratedProjectView
	}

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