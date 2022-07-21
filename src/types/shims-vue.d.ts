// https://next.vuex.vuejs.org/guide/migrating-to-4-0-from-3-x.html#typescript-support
import { Store } from 'vuex'

import type {
	RootStoreState,
	AttachmentState,
	AuthState,
	ConfigState,
	KanbanState,
	LabelState,
	ListState,
	NamespaceState,
	TaskState,
} from '@/store/types'

declare module '@vue/runtime-core' {

  interface ComponentCustomProperties {
    $store: Store<RootStoreState & {
			config: ConfigState,
			auth: AuthState,
			namespaces: NamespaceState,
			kanban: KanbanState,
			tasks: TaskState,
			lists: ListState,
			attachments: AttachmentState,
			labels: LabelState,
		}>
  }
}