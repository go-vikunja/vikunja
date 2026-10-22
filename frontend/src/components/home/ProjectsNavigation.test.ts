import {describe, it, expect, vi} from 'vitest'
import {mount} from '@vue/test-utils'
import {defineComponent, h, nextTick, type PropType} from 'vue'
import draggable from 'zhyswan-vuedraggable'

import ProjectsNavigation from './ProjectsNavigation.vue'
import type {ProjectResponse} from '@/client/queries/projects'

vi.mock('@/client/queries/projects', () => ({
	useUpdateProjectMutation: () => ({mutateAsync: vi.fn()}),
}))
vi.mock('@/composables/useProjects', () => ({
	useProjects: () => ({projects: {}}),
}))

const ProjectsNavigationItem = defineComponent({
	props: {project: {type: Object as PropType<ProjectResponse>, required: true}},
	setup: props => () => h('li', {'data-project-id': props.project.id}),
})

function projects(...ids: number[]) {
	return ids.map(id => ({id, position: id}) as ProjectResponse)
}

function renderedIds(wrapper: ReturnType<typeof mount>) {
	return wrapper.findAll('li').map(li => Number(li.attributes('data-project-id')))
}

describe('ProjectsNavigation', () => {
	it('does not patch the list while Sortable owns its DOM', async () => {
		const errors: unknown[] = []
		const wrapper = mount(ProjectsNavigation, {
			props: {modelValue: projects(1, 2, 3, 4), canEditOrder: true},
			attachTo: document.body,
			global: {
				stubs: {ProjectsNavigationItem},
				config: {errorHandler: error => { errors.push(error) }},
			},
		})
		const list = wrapper.findComponent(draggable)

		list.vm.$emit('start')
		await nextTick()

		// Sortable moves the dragged node while it hovers over another list.
		const dragged = wrapper.find('li[data-project-id="2"]').element
		const originalParent = dragged.parentElement!
		const otherList = document.createElement('menu')
		document.body.appendChild(otherList)
		otherList.appendChild(dragged)

		// A refetch landing mid-drag moves 4 in front of the dragged node.
		await wrapper.setProps({modelValue: projects(1, 4, 2, 3)})

		expect(errors).toEqual([])

		// vuedraggable puts the node back on drop.
		originalParent.insertBefore(dragged, originalParent.children[1] ?? null)
		list.vm.$emit('end', {})
		await nextTick()
		await nextTick()

		expect(errors).toEqual([])
		expect(renderedIds(wrapper)).toEqual([1, 4, 2, 3])

		wrapper.unmount()
		otherList.remove()
	})
})
