import {createPinia} from 'pinia'
import {afterEach, beforeEach, expect, it, vi} from 'vitest'
import {ref} from 'vue'
import {mount, flushPromises, enableAutoUnmount} from '@vue/test-utils'
import {QueryClient, VueQueryPlugin} from '@tanstack/vue-query'
import {taskKeys} from '@/client/queries/tasks'
import TaskLinkPill from './TaskLinkPill.vue'
import {taskLinkCurrentProjectIdKey} from './taskLinkContext'
const sdk = vi.hoisted(() => ({tasksRead: vi.fn()}))
vi.mock('@/client/generated', () => sdk)
const baseStore = vi.hoisted(() => ({currentProjectId: 0}))
vi.mock('@/stores/base', () => ({useBaseStore: () => baseStore}))
const projectList = vi.hoisted(() => ({projects: {} as Record<number, {id: number, title: string}>}))
vi.mock('@/composables/useProjects', () => ({useProjects: () => projectList}))
enableAutoUnmount(afterEach)
let client: QueryClient
beforeEach(() => {
	client = new QueryClient({defaultOptions: {queries: {retry: false, staleTime: 60000}}})
	sdk.tasksRead.mockReset()
	baseStore.currentProjectId = 0
	projectList.projects = {}
})
const HREF = 'http://localhost:3000/tasks/5'
function task(overrides: Record<string, unknown> = {}) {
	return {id: 5, title: 'Fix the thing', identifier: 'FT-12', index: 12, done: false, project_id: 1, ...overrides}
}
function pill(id = 5, provide: Record<symbol, unknown> = {}) {
	return mount(TaskLinkPill, {
		props: {href: `http://localhost:3000/tasks/${id}`},
		global: {provide, plugins: [createPinia(), [VueQueryPlugin, {queryClient: client}]], mocks: {$t: (key: string) => key}, stubs: {TaskGlanceTooltip: {template: '<span><slot /></span>'}, Icon: true}},
	})
}
it('updates an already mounted link from the task cache', async () => {
	client.setQueryData(taskKeys.detail(5), {id: 5, title: 'Original', index: 12})
	const wrapper = pill()
	expect(wrapper.text()).toContain('Original')
	client.setQueryData(taskKeys.detail(5), {id: 5, title: 'Renamed', index: 12, done: true})
	await flushPromises()
	expect(wrapper.text()).toContain('Renamed')
	expect(wrapper.find('.task-link-pill--done').exists()).toBe(true)
})
it('discards the old task when the href changes during a request', async () => {
	let resolveFirst!: (value: unknown) => void
	sdk.tasksRead.mockReturnValueOnce(new Promise(resolve => { resolveFirst = resolve })).mockResolvedValueOnce({data: {id: 6, title: 'Second'}})
	const wrapper = pill()
	await flushPromises()
	await wrapper.setProps({href: 'http://localhost:3000/tasks/6'})
	await flushPromises()
	resolveFirst({data: {id: 5, title: 'Stale'}})
	await flushPromises()
	expect(wrapper.text()).toContain('Second')
	expect(wrapper.text()).not.toContain('Stale')
})
it('keeps loaded data when a refresh fails', async () => {
	client.setQueryData(taskKeys.detail(5), {id: 5, title: 'Original'})
	const wrapper = pill()
	sdk.tasksRead.mockRejectedValue(new Error('offline'))
	await client.invalidateQueries({queryKey: taskKeys.detail(5)})
	await flushPromises()
	expect(wrapper.text()).toContain('Original')
})
it('shows the href while loading', () => {
	sdk.tasksRead.mockReturnValue(new Promise(() => {}))
	expect(pill().find('.task-link-pill--loading').text()).toBe(HREF)
})
it('renders identifier and title once loaded', async () => {
	sdk.tasksRead.mockResolvedValue({data: task()})
	const wrapper = pill()
	await flushPromises()
	expect(wrapper.find('.task-link-pill__identifier').text()).toBe('FT-12')
	expect(wrapper.find('.task-link-pill__title').text()).toBe('Fix the thing')
	expect(wrapper.find('.task-link-pill--done').exists()).toBe(false)
})
it('falls back to #index when the task has no identifier', async () => {
	sdk.tasksRead.mockResolvedValue({data: task({identifier: ''})})
	const wrapper = pill()
	await flushPromises()
	expect(wrapper.find('.task-link-pill__identifier').text()).toBe('#12')
})
it('marks done tasks visually and for screen readers', async () => {
	sdk.tasksRead.mockResolvedValue({data: task({done: true})})
	const wrapper = pill()
	await flushPromises()
	expect(wrapper.find('.task-link-pill--done').exists()).toBe(true)
	expect(wrapper.find('.task-link-pill--done .is-sr-only').text()).toBe('task.attributes.done')
})
it('shows the project name when the task lives in another project', async () => {
	projectList.projects = {2: {id: 2, title: 'Other project'}}
	baseStore.currentProjectId = 1
	sdk.tasksRead.mockResolvedValue({data: task({project_id: 2})})
	const wrapper = pill()
	await flushPromises()
	expect(wrapper.find('.task-link-pill__project').text()).toBe('Other project ›')
})
it('hides the project name when the task is in the current project', async () => {
	projectList.projects = {1: {id: 1, title: 'Current'}}
	baseStore.currentProjectId = 1
	sdk.tasksRead.mockResolvedValue({data: task({project_id: 1})})
	const wrapper = pill()
	await flushPromises()
	expect(wrapper.find('.task-link-pill__project').exists()).toBe(false)
})
it('prefers the project id provided by the surrounding view', async () => {
	projectList.projects = {1: {id: 1, title: 'Current'}, 2: {id: 2, title: 'Viewed'}}
	baseStore.currentProjectId = 1
	sdk.tasksRead.mockResolvedValue({data: task({project_id: 2})})
	const wrapper = pill(5, {[taskLinkCurrentProjectIdKey as symbol]: ref(2)})
	await flushPromises()
	expect(wrapper.find('.task-link-pill__project').exists()).toBe(false)
})
it('renders a plain anchor when the fetch fails', async () => {
	sdk.tasksRead.mockRejectedValue(new Error('nope'))
	const wrapper = pill()
	await flushPromises()
	const link = wrapper.find('a.task-link-pill--fallback')
	expect(link.attributes('href')).toBe(HREF)
	expect(link.text()).toBe(HREF)
})
it('renders the href as text without an anchor when it is not a task url', async () => {
	const wrapper = mount(TaskLinkPill, {
		props: {href: 'https://example.com/'},
		global: {plugins: [createPinia(), [VueQueryPlugin, {queryClient: client}]], mocks: {$t: (key: string) => key}, stubs: {TaskGlanceTooltip: {template: '<span><slot /></span>'}, Icon: true}},
	})
	await flushPromises()
	expect(wrapper.find('a').exists()).toBe(false)
	expect(wrapper.find('span').text()).toBe('https://example.com/')
	expect(sdk.tasksRead).not.toHaveBeenCalled()
})
it('emits open with the task instead of following the anchor', async () => {
	sdk.tasksRead.mockResolvedValue({data: task()})
	const wrapper = pill()
	await flushPromises()
	await wrapper.find('a.task-link-pill--task').trigger('click')
	expect(wrapper.emitted('open')?.[0]?.[0]).toMatchObject({id: 5})
})
it('lets modified clicks through so the link can open in a new tab', async () => {
	sdk.tasksRead.mockResolvedValue({data: task()})
	const wrapper = pill()
	await flushPromises()
	await wrapper.find('a.task-link-pill--task').trigger('click', {ctrlKey: true})
	expect(wrapper.emitted('open')).toBeUndefined()
})
