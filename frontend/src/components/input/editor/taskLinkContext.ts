import type {InjectionKey, Ref} from 'vue'

// The global project selection can point at the backdrop project in task modals.
export const taskLinkCurrentProjectIdKey: InjectionKey<Ref<number | undefined>> = Symbol('taskLinkCurrentProjectId')
