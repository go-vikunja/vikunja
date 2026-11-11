import {nextTick, ref, type Ref} from 'vue'
import {flushPromises, shallowMount} from '@vue/test-utils'
import {QueryClient, VueQueryPlugin} from '@tanstack/vue-query'
import {beforeEach, describe, expect, it, vi} from 'vitest'

import type {Project, ProjectView} from '@/client/generated'

const state = vi.hoisted(() => ({
	project: undefined as Ref<Project> | undefined,
}))

const transport = vi.hoisted(() => ({
	createProjectView: vi.fn(),
	deleteProjectView: vi.fn(),
	updateProjectView: vi.fn(),
}))

// Mutations invalidate the project, so a reload has to answer with the same views the
// cache was seeded with instead of dropping them.
const server = vi.hoisted(() => ({viewsByProject: {} as Record<number, unknown[]>}))

vi.mock('@/composables/useProject', async () => {
	const {ref} = await import('vue')
	return {
		useProject: () => ({
			project: state.project,
			isLoading: ref(false),
			error: ref(null),
		}),
	}
})

vi.mock('@/client/generated', async importOriginal => ({
	...await importOriginal<typeof import('@/client/generated')>(),
	projectsRead: async ({path}: {path: {id: number}}) => ({
		data: {id: path.id, views: server.viewsByProject[path.id] ?? []},
	}),
	projectViewsCreate: async (input: unknown) => ({data: await transport.createProjectView(input)}),
	projectViewsDelete: (input: unknown) => transport.deleteProjectView(input),
	projectViewsUpdate: async (input: unknown) => ({data: await transport.updateProjectView(input)}),
}))

vi.mock('@/message', () => ({error: vi.fn(), success: vi.fn()}))
vi.mock('vue-i18n', async importOriginal => ({
	...await importOriginal<typeof import('vue-i18n')>(),
	useI18n: () => ({t: (key: string) => key}),
}))

import ProjectSettingsViews from './ProjectSettingsViews.vue'
import {normalizeProject, projectKeys} from '@/client/queries/projects'
import {error, success} from '@/message'

const {createProjectView, deleteProjectView, updateProjectView} = transport

type ProjectSettingsViewsVm = {
	showCreateForm: boolean
	newView: ProjectView & {project_id?: number}
	viewIdToDelete: number | null
	showDeleteModal: boolean
	viewToEdit: ProjectView | null
	views: ProjectView[]
	isSaving: boolean
	createView: () => Promise<void>
	deleteView: (viewId: number | null) => Promise<void>
	saveView: (view: ProjectView) => Promise<void>
	saveViewPosition: (event: {newIndex: number}) => Promise<void>
}

function deferred<T>() {
	let resolve!: (value: T) => void
	let reject!: (reason: unknown) => void
	const promise = new Promise<T>((resolvePromise, rejectPromise) => {
		resolve = resolvePromise
		reject = rejectPromise
	})
	return {promise, resolve, reject}
}

let client: QueryClient

// The views the component renders come from the project cache, the same place the
// mutations write to.
function seedViews(views: ProjectView[] = [], projectId = 1) {
	server.viewsByProject[projectId] = views
	client.setQueryData(projectKeys.detail(projectId), normalizeProject({id: projectId, views}))
}

function mountView(projectId = 1) {
	return shallowMount(ProjectSettingsViews, {
		props: {projectId},
		global: {
			plugins: [[VueQueryPlugin, {queryClient: client}]],
			mocks: {$t: (key: string) => key},
			stubs: {
				CreateEdit: {template: '<div><slot /></div>'},
				Icon: true,
				Modal: true,
				draggable: true,
			},
		},
	})
}

describe('ProjectSettingsViews', () => {
	beforeEach(() => {
		client = new QueryClient({defaultOptions: {queries: {staleTime: Infinity, retry: false}}})
		server.viewsByProject = {}
		seedViews()
		state.project = ref({id: 1, max_permission: 2})
		vi.clearAllMocks()
	})

	it('follows project cache updates instead of snapshotting the views', async () => {
		const wrapper = mountView()
		const vm = wrapper.vm as unknown as ProjectSettingsViewsVm

		seedViews([{id: 10, project_id: 1, title: 'Created elsewhere', view_kind: 'list'}])
		await flushPromises()
		await nextTick()

		expect(vm.views.map(view => view.id)).toEqual([10])
	})

	it('resets route-scoped editor state when the project changes', async () => {
		const wrapper = mountView()
		const vm = wrapper.vm as unknown as ProjectSettingsViewsVm
		const staleView: ProjectView = {
			id: 10,
			project_id: 1,
			title: 'Stale view',
			view_kind: 'list',
		}
		const staleDraft = vm.newView

		vm.showCreateForm = true
		vm.newView.title = 'Stale draft'
		vm.viewToEdit = staleView
		vm.viewIdToDelete = staleView.id ?? null
		vm.showDeleteModal = true

		await wrapper.setProps({projectId: 2})

		expect(vm.showCreateForm).toBe(false)
		expect(vm.newView).not.toBe(staleDraft)
		expect(vm.newView).toMatchObject({project_id: 2, title: ''})
		expect(vm.viewToEdit).toBeNull()
		expect(vm.viewIdToDelete).toBeNull()
		expect(vm.showDeleteModal).toBe(false)
	})

	it('does not reset the new project draft when an old create finishes', async () => {
		const create = deferred<ProjectView>()
		vi.mocked(createProjectView).mockReturnValue(create.promise)
		const wrapper = mountView()
		const vm = wrapper.vm as unknown as ProjectSettingsViewsVm

		vm.showCreateForm = true
		vm.newView.title = 'Project A draft'
		const pendingCreate = vm.createView()
		await vi.waitFor(() => expect(createProjectView).toHaveBeenCalledOnce())

		await wrapper.setProps({projectId: 2})
		vm.showCreateForm = true
		vm.newView.title = 'Project B draft'
		const projectBDraft = vm.newView

		create.resolve({id: 10, project_id: 1, title: 'Project A view', view_kind: 'list'})
		await pendingCreate

		expect(vm.showCreateForm).toBe(true)
		expect(vm.newView).toBe(projectBDraft)
		expect(vm.newView.title).toBe('Project B draft')
		expect(success).not.toHaveBeenCalled()
	})

	it('does not close the new project modal when an old delete finishes', async () => {
		const remove = deferred<void>()
		vi.mocked(deleteProjectView).mockReturnValue(remove.promise)
		const wrapper = mountView()
		const vm = wrapper.vm as unknown as ProjectSettingsViewsVm

		vm.showDeleteModal = true
		vm.viewIdToDelete = 10
		const pendingDelete = vm.deleteView(10)
		await vi.waitFor(() => expect(deleteProjectView).toHaveBeenCalledOnce())

		await wrapper.setProps({projectId: 2})
		vm.showDeleteModal = true
		vm.viewIdToDelete = 20

		remove.resolve()
		await pendingDelete

		expect(vm.showDeleteModal).toBe(true)
		expect(vm.viewIdToDelete).toBe(20)
	})

	it('suppresses an old create error without clearing a new project mutation', async () => {
		const projectACreate = deferred<ProjectView>()
		const projectBSave = deferred<ProjectView>()
		vi.mocked(createProjectView).mockReturnValue(projectACreate.promise)
		vi.mocked(updateProjectView).mockReturnValue(projectBSave.promise)
		const wrapper = mountView()
		const vm = wrapper.vm as unknown as ProjectSettingsViewsVm

		vm.showCreateForm = true
		vm.newView.title = 'Project A draft'
		const pendingProjectACreate = vm.createView()
		await vi.waitFor(() => expect(createProjectView).toHaveBeenCalledOnce())

		await wrapper.setProps({projectId: 2})
		const projectBView: ProjectView = {
			id: 20,
			project_id: 2,
			title: 'Project B view',
			view_kind: 'list',
		}
		vm.viewToEdit = projectBView
		const pendingProjectBSave = vm.saveView(projectBView)
		await vi.waitFor(() => expect(updateProjectView).toHaveBeenCalledOnce())

		projectACreate.reject(new Error('Project A failed'))
		await pendingProjectACreate

		expect(error).not.toHaveBeenCalled()
		expect(vm.viewToEdit).toEqual(projectBView)
		expect(vm.isSaving).toBe(true)

		projectBSave.resolve(projectBView)
		await pendingProjectBSave
	})

	it('keeps the new project edit pending when an old save finishes', async () => {
		const projectASave = deferred<ProjectView>()
		const projectBSave = deferred<ProjectView>()
		vi.mocked(updateProjectView)
			.mockReturnValueOnce(projectASave.promise)
			.mockReturnValueOnce(projectBSave.promise)
		const wrapper = mountView()
		const vm = wrapper.vm as unknown as ProjectSettingsViewsVm
		const projectAView: ProjectView = {
			id: 10,
			project_id: 1,
			title: 'Project A view',
			view_kind: 'list',
		}
		const projectBView: ProjectView = {
			id: 20,
			project_id: 2,
			title: 'Project B view',
			view_kind: 'list',
		}

		vm.viewToEdit = projectAView
		const pendingProjectASave = vm.saveView(projectAView)
		await vi.waitFor(() => expect(updateProjectView).toHaveBeenCalledOnce())

		await wrapper.setProps({projectId: 2})
		vm.viewToEdit = projectBView
		const pendingProjectBSave = vm.saveView(projectBView)
		await vi.waitFor(() => expect(updateProjectView).toHaveBeenCalledTimes(2))

		projectASave.resolve(projectAView)
		await pendingProjectASave

		expect(vm.viewToEdit).toEqual(projectBView)
		expect(vm.isSaving).toBe(true)
		expect(success).not.toHaveBeenCalled()

		projectBSave.resolve(projectBView)
		await pendingProjectBSave

		expect(vm.viewToEdit).toBeNull()
		expect(vm.isSaving).toBe(false)
		expect(success).toHaveBeenCalledOnce()
	})

	it('suppresses an old delete error after the project changes', async () => {
		const remove = deferred<void>()
		vi.mocked(deleteProjectView).mockReturnValue(remove.promise)
		const wrapper = mountView()
		const vm = wrapper.vm as unknown as ProjectSettingsViewsVm
		const pendingDelete = vm.deleteView(10)
		await vi.waitFor(() => expect(deleteProjectView).toHaveBeenCalledOnce())

		await wrapper.setProps({projectId: 2})
		remove.reject(new Error('Project A delete failed'))
		await pendingDelete

		expect(error).not.toHaveBeenCalled()
	})

	it('suppresses an old save error after the project changes', async () => {
		const save = deferred<ProjectView>()
		vi.mocked(updateProjectView).mockReturnValue(save.promise)
		const wrapper = mountView()
		const vm = wrapper.vm as unknown as ProjectSettingsViewsVm
		const view: ProjectView = {
			id: 10,
			project_id: 1,
			title: 'Project A view',
			view_kind: 'list',
		}
		const pendingSave = vm.saveView(view)
		await vi.waitFor(() => expect(updateProjectView).toHaveBeenCalledOnce())

		await wrapper.setProps({projectId: 2})
		save.reject(new Error('Project A save failed'))
		await pendingSave

		expect(error).not.toHaveBeenCalled()
	})

	it('suppresses an old reorder error after the project changes', async () => {
		const view: ProjectView = {
			id: 10,
			project_id: 1,
			title: 'Project A view',
			view_kind: 'list',
		}
		seedViews([view])
		const save = deferred<ProjectView>()
		vi.mocked(updateProjectView).mockReturnValue(save.promise)
		const wrapper = mountView()
		const vm = wrapper.vm as unknown as ProjectSettingsViewsVm
		await flushPromises()
		const pendingSave = vm.saveViewPosition({newIndex: 0})
		await vi.waitFor(() => expect(updateProjectView).toHaveBeenCalledOnce())

		await wrapper.setProps({projectId: 2})
		save.reject(new Error('Project A reorder failed'))
		await pendingSave

		expect(error).not.toHaveBeenCalled()
	})

	it('reports current-project delete, save, and reorder errors', async () => {
		const view: ProjectView = {
			id: 10,
			project_id: 1,
			title: 'Project A view',
			view_kind: 'list',
		}
		seedViews([view])
		const removeError = new Error('Delete failed')
		const saveError = new Error('Save failed')
		const reorderError = new Error('Reorder failed')
		vi.mocked(deleteProjectView).mockRejectedValueOnce(removeError)
		vi.mocked(updateProjectView)
			.mockRejectedValueOnce(saveError)
			.mockRejectedValueOnce(reorderError)
		const wrapper = mountView()
		const vm = wrapper.vm as unknown as ProjectSettingsViewsVm
		await flushPromises()

		await vm.deleteView(10)
		await vm.saveView(view)
		await vm.saveViewPosition({newIndex: 0})

		expect(error).toHaveBeenNthCalledWith(1, removeError)
		expect(error).toHaveBeenNthCalledWith(2, saveError)
		expect(error).toHaveBeenNthCalledWith(3, reorderError)
	})
})
