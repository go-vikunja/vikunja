import {describe, it, expect, vi} from 'vitest'
import {mount} from '@vue/test-utils'
import {createI18n} from 'vue-i18n'
import ProjectInfo from './ProjectInfo.vue'

const projects = vi.hoisted(() => ({value: {} as Record<number, {id: number, description: string}>}))

vi.mock('@/composables/useProjects', () => ({
	useProjects: () => ({get projects() { return projects.value }}),
}))
vi.mock('@/composables/useGoBack', () => ({useGoBack: () => vi.fn()}))

function mountInfo(projectId: number, favoriteDescription: string) {
	const i18n = createI18n({
		legacy: false,
		locale: 'en',
		messages: {en: {project: {favoriteDescription, noDescriptionAvailable: 'none'}}},
	})
	return mount(ProjectInfo, {
		props: {projectId},
		global: {
			plugins: [i18n],
			stubs: {
				Modal: {template: '<div><slot /></div>'},
				Card: {template: '<div><slot /></div>'},
			},
		},
	})
}

describe('ProjectInfo', () => {
	it('renders the favorites description as text', () => {
		projects.value = {[-1]: {id: -1, description: 'backend text'}}
		const wrapper = mountInfo(-1, '<img src="x"> Favorites')

		expect(wrapper.find('img').exists()).toBe(false)
		expect(wrapper.text()).toContain('<img src="x"> Favorites')
	})

	it('renders sanitized project descriptions as html', () => {
		projects.value = {[1]: {id: 1, description: '<p><strong>bold</strong></p>'}}
		const wrapper = mountInfo(1, 'unused')

		expect(wrapper.find('strong').text()).toBe('bold')
	})
})
