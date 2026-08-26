import {describe, it, expect, beforeEach, vi} from 'vitest'
import {createPinia, setActivePinia} from 'pinia'
import {ref} from 'vue'
import {mount, flushPromises} from '@vue/test-utils'

const fetchTaskById = vi.fn()
vi.mock('@/helpers/fetchTaskById', () => ({
	fetchTaskById: (id: number) => fetchTaskById(id),
}))

vi.mock('@/helpers/taskCache', async () => {
	const vue = await import('vue')
	return {
		taskCacheVersion: vue.ref(0),
		taskCacheIdentityVersion: vue.ref(0),
	}
})

const baseStore = {currentProject: null as {id: number} | null}
vi.mock('@/stores/base', () => ({
	useBaseStore: () => baseStore,
}))

const projectStore = {projects: {} as Record<number, {id: number, title: string}>}
vi.mock('@/stores/projects', () => ({
	useProjectStore: () => projectStore,
}))

vi.mock('@/helpers/getProjectTitle', () => ({
	getProjectTitle: (project: {title: string}) => project.title,
}))

import {taskCacheVersion, taskCacheIdentityVersion} from '@/helpers/taskCache'
import TaskLinkPill from './TaskLinkPill.vue'
import {taskLinkCurrentProjectIdKey} from './taskLinkContext'

const HREF = 'http://localhost:3000/tasks/5'

function makeTask(overrides: Record<string, unknown> = {}) {
	return {
		id: 5,
		title: 'Fix the thing',
		identifier: 'FT-12',
		index: 12,
		done: false,
		projectId: 1,
		...overrides,
	}
}

async function mountPill(props: Partial<{href: string}> = {}, provide: Record<symbol, unknown> = {}) {
	const wrapper = mount(TaskLinkPill, {
		props: {
			href: HREF,
			...props,
		},
		global: {
			provide,
			mocks: {$t: (key: string) => key},
			stubs: {
				TaskGlanceTooltip: {template: '<span><slot /></span>'},
				Icon: {template: '<i />'},
			},
		},
	})
	await flushPromises()
	return wrapper
}

describe('TaskLinkPill', () => {
	beforeEach(() => {
		setActivePinia(createPinia())
		baseStore.currentProject = null
		projectStore.projects = {}
		fetchTaskById.mockReset()
	})

	it('shows the href while loading', () => {
		fetchTaskById.mockReturnValue(new Promise(() => {}))
		const wrapper = mount(TaskLinkPill, {props: {href: HREF}})

		expect(wrapper.find('.task-link-pill--loading').text()).toBe(HREF)
	})

	it('renders identifier and title once loaded', async () => {
		fetchTaskById.mockResolvedValue(makeTask())
		const wrapper = await mountPill()

		expect(wrapper.find('.task-link-pill__identifier').text()).toBe('FT-12')
		expect(wrapper.find('.task-link-pill__title').text()).toBe('Fix the thing')
		expect(wrapper.find('.task-link-pill--done').exists()).toBe(false)
	})

	it('falls back to #index when the task has no identifier', async () => {
		fetchTaskById.mockResolvedValue(makeTask({identifier: ''}))
		const wrapper = await mountPill()

		expect(wrapper.find('.task-link-pill__identifier').text()).toBe('#12')
	})

	it('marks done tasks visually and for screen readers', async () => {
		fetchTaskById.mockResolvedValue(makeTask({done: true}))
		const wrapper = await mountPill()

		expect(wrapper.find('.task-link-pill--done').exists()).toBe(true)
		expect(wrapper.find('.task-link-pill--done .is-sr-only').text()).toBe('task.attributes.done')
	})

	it('shows the project name when the task lives in another project', async () => {
		projectStore.projects = {2: {id: 2, title: 'Other project'}}
		baseStore.currentProject = {id: 1}
		fetchTaskById.mockResolvedValue(makeTask({projectId: 2}))

		const wrapper = await mountPill()

		expect(wrapper.find('.task-link-pill__project').text()).toBe('Other project ›')
	})

	it('hides the project name when the task is in the current project', async () => {
		projectStore.projects = {1: {id: 1, title: 'Current'}}
		baseStore.currentProject = {id: 1}
		fetchTaskById.mockResolvedValue(makeTask({projectId: 1}))

		const wrapper = await mountPill()

		expect(wrapper.find('.task-link-pill__project').exists()).toBe(false)
	})

	it('prefers the project id provided by the surrounding view', async () => {
		projectStore.projects = {1: {id: 1, title: 'Current'}, 2: {id: 2, title: 'Viewed'}}
		baseStore.currentProject = {id: 1}
		fetchTaskById.mockResolvedValue(makeTask({projectId: 2}))

		const wrapper = await mountPill({}, {[taskLinkCurrentProjectIdKey as symbol]: ref(2)})

		expect(wrapper.find('.task-link-pill__project').exists()).toBe(false)
	})

	it('renders a plain anchor when the fetch fails', async () => {
		fetchTaskById.mockRejectedValue(new Error('nope'))
		const wrapper = await mountPill()

		const link = wrapper.find('a.task-link-pill--fallback')
		expect(link.exists()).toBe(true)
		expect(link.attributes('href')).toBe(HREF)
		expect(link.text()).toBe(HREF)
	})

	it('renders the href as text without an anchor when it is not a task url', async () => {
		const wrapper = await mountPill({href: 'https://example.com/'})

		expect(wrapper.find('a').exists()).toBe(false)
		expect(wrapper.find('span').text()).toBe('https://example.com/')
		expect(fetchTaskById).not.toHaveBeenCalled()
	})

	it('emits open with the task instead of following the anchor', async () => {
		fetchTaskById.mockResolvedValue(makeTask())
		const wrapper = await mountPill()

		await wrapper.find('a.task-link-pill--task').trigger('click')

		expect(wrapper.emitted('open')?.[0]?.[0]).toMatchObject({id: 5})
	})

	it('lets modified clicks through so the link can open in a new tab', async () => {
		fetchTaskById.mockResolvedValue(makeTask())
		const wrapper = await mountPill()

		await wrapper.find('a.task-link-pill--task').trigger('click', {ctrlKey: true})

		expect(wrapper.emitted('open')).toBeUndefined()
	})

	it('refetches when the cache is invalidated', async () => {
		fetchTaskById.mockResolvedValue(makeTask())
		const wrapper = await mountPill()

		fetchTaskById.mockClear()
		fetchTaskById.mockResolvedValue(makeTask({title: 'Renamed'}))
		taskCacheVersion.value++
		await flushPromises()

		expect(fetchTaskById).toHaveBeenCalledWith(5)
		expect(wrapper.find('.task-link-pill__title').text()).toBe('Renamed')
	})

	it('keeps the task when a refetch fails', async () => {
		fetchTaskById.mockResolvedValue(makeTask())
		const wrapper = await mountPill()

		fetchTaskById.mockRejectedValue(new Error('nope'))
		taskCacheVersion.value++
		await flushPromises()

		expect(wrapper.find('.task-link-pill--fallback').exists()).toBe(false)
		expect(wrapper.find('.task-link-pill__title').text()).toBe('Fix the thing')
	})

	it('drops the task without refetching when the identity changes', async () => {
		fetchTaskById.mockResolvedValue(makeTask())
		const wrapper = await mountPill()

		fetchTaskById.mockClear()
		taskCacheIdentityVersion.value++
		await flushPromises()

		expect(fetchTaskById).not.toHaveBeenCalled()
		expect(wrapper.find('.task-link-pill--loading').text()).toBe(HREF)

		fetchTaskById.mockResolvedValue(makeTask({title: 'Under new auth'}))
		taskCacheVersion.value++
		await flushPromises()

		expect(wrapper.find('.task-link-pill__title').text()).toBe('Under new auth')
	})

	it('discards a result superseded by a newer fetch', async () => {
		let resolveFirst: (task: unknown) => void = () => {}
		fetchTaskById.mockReturnValueOnce(new Promise(resolve => {
			resolveFirst = resolve
		}))
		const wrapper = await mountPill()

		fetchTaskById.mockResolvedValue(makeTask({title: 'Renamed'}))
		taskCacheVersion.value++
		await flushPromises()

		resolveFirst(makeTask({title: 'Stale'}))
		await flushPromises()

		expect(wrapper.find('.task-link-pill__title').text()).toBe('Renamed')
	})
})
