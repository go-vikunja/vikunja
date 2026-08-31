import type {InjectionKey, Ref} from 'vue'
import type {IProject} from '@/modelTypes/IProject'

// TipTap provides the host task's project (its `projectId`) so pills can hide the
// prefix for same-project tasks; baseStore.currentProject is stale on direct task loads
// and points at the backdrop project in modals.
export const taskLinkCurrentProjectIdKey: InjectionKey<Ref<IProject['id'] | undefined>> = Symbol('taskLinkCurrentProjectId')
