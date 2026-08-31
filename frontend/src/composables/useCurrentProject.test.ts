import {defineComponent, effectScope, h, ref} from 'vue'
import {mount} from '@vue/test-utils'
import {VueQueryPlugin} from '@tanstack/vue-query'
import {beforeEach, describe, expect, it, vi} from 'vitest'

import {queryClient} from '@/client/queryClient'
import {projectKeys} from '@/client/queries/projects'

const sdk = vi.hoisted(() => ({
	projectsList: vi.fn(),
	projectsRead: vi.fn(),
}))
const useBaseStore = vi.hoisted(() => vi.fn())

vi.mock('@/client/generated', () => sdk)
vi.mock('@/stores/base', () => ({useBaseStore}))

import {useCurrentProject} from './useCurrentProject'

async function mountCurrentProject(projectId: number) {
	useBaseStore.mockReturnValue({currentProjectId: ref(projectId)})
	const component = defineComponent({
		setup() {
			useCurrentProject()
			return () => h('div')
		},
	})
	const wrapper = mount(component, {
		global: {
			plugins: [[VueQueryPlugin, {queryClient}]],
		},
	})

	return {queryClient, wrapper}
}

describe('useCurrentProject', () => {
	beforeEach(() => {
		queryClient.clear()
		Object.values(sdk).forEach(mock => mock.mockReset())
		sdk.projectsRead.mockResolvedValue({data: {id: 42, title: 'Project'}})
		sdk.projectsList.mockResolvedValue({data: {items: [], total_pages: 1}})
	})

	it('can be created outside a Vue injection context', () => {
		useBaseStore.mockReturnValue({currentProjectId: ref(0)})
		const scope = effectScope()

		expect(() => scope.run(() => useCurrentProject())).not.toThrow()
		scope.stop()
	})

	it('activates only the detail query for a real project', async () => {
		const {queryClient, wrapper} = await mountCurrentProject(42)

		await vi.waitFor(() => expect(sdk.projectsRead).toHaveBeenCalledOnce())

		expect(sdk.projectsList).not.toHaveBeenCalled()
		expect(queryClient.getQueryCache().find({queryKey: projectKeys.detail(42)})?.isActive()).toBe(true)
		expect(queryClient.getQueryCache().find({queryKey: projectKeys.list()})?.isActive()).toBe(false)
		wrapper.unmount()
	})

	it.each([-1, -2])('activates only the navigation query for pseudo-project %i', async projectId => {
		const {queryClient, wrapper} = await mountCurrentProject(projectId)

		await vi.waitFor(() => expect(sdk.projectsList).toHaveBeenCalledOnce())

		expect(sdk.projectsRead).not.toHaveBeenCalled()
		expect(queryClient.getQueryCache().find({queryKey: projectKeys.detail(projectId)})?.isActive()).toBe(false)
		expect(queryClient.getQueryCache().find({queryKey: projectKeys.list()})?.isActive()).toBe(true)
		wrapper.unmount()
	})
})
