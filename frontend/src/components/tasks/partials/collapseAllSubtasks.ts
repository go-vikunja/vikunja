import type {InjectionKey, Ref} from 'vue'

export interface CollapseAllSubtasksState {
	collapsed: boolean
	// Bumped on every toggle so a task the user expanded individually is pulled back in
	// sync even when `collapsed` itself does not change.
	token: number
}

export const collapseAllSubtasksKey: InjectionKey<Ref<CollapseAllSubtasksState>> = Symbol('collapseAllSubtasks')
