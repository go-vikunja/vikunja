import {describe, it, expect, beforeEach, vi} from 'vitest'
import {mount} from '@vue/test-utils'
import {createI18n} from 'vue-i18n'
import {ref, reactive, computed} from 'vue'

import Home from './Home.vue'
import en from '@/i18n/lang/en.json'

const mockGetHistory = vi.fn(() => [{id: 1}, {id: 2}])

vi.mock('vue-router', () => ({
	useRoute: () => ({query: {}}),
	useRouter: () => ({push: vi.fn()}),
	createRouter: () => ({
		push: vi.fn(),
		resolve: vi.fn(() => ({href: '/'})),
		currentRoute: {value: {query: {}}},
		beforeEach: vi.fn(),
		afterEach: vi.fn(),
		install: vi.fn(),
	}),
	createWebHistory: vi.fn(),
	RouterLink: {template: '<a><slot /></a>'},
}))

vi.mock('@/modules/projectHistory', () => ({
	getHistory: (...args: any[]) => mockGetHistory(...args),
}))

vi.mock('@/composables/useDaytimeSalutation', () => ({
	useDaytimeSalutation: () => computed(() => 'Hello!'),
}))

vi.mock('@/helpers/time/formatDate', () => ({
	formatDateSince: () => '',
	formatDisplayDate: () => '',
}))

vi.mock('@/helpers/parseDateOrNull', () => ({
	parseDateOrNull: () => null,
}))

const mockProjects = reactive<Record<number, any>>({})
const mockAuthenticated = ref(true)
const mockShowLastViewed = ref(true)
const mockHasProjects = ref(false)

vi.mock('@/stores/auth', () => ({
	useAuthStore: () => ({
		get authenticated() {
			return mockAuthenticated.value
		},
		info: null,
		settings: {
			frontendSettings: {
				get showLastViewed() {
					return mockShowLastViewed.value
				},
			},
		},
	}),
}))

vi.mock('@/stores/projects', () => ({
	useProjectStore: () => ({
		projects: mockProjects,
		hasProjects: mockHasProjects,
	}),
}))

const i18n = createI18n({legacy: false, locale: 'en', messages: {en}})

function mountHome() {
	return mount(Home, {
		global: {
			plugins: [i18n],
			stubs: {
				Message: true,
				AddTask: true,
				ImportHint: true,
				ShowTasks: true,
				ProjectCardGrid: {
					template: '<div class="project-card-grid-stub" />',
					props: ['projects', 'showEvenNumberOfProjects'],
				},
				RouterLink: true,
			},
		},
	})
}

function seedProjects(...ids: number[]) {
	for (const id of ids) {
		mockProjects[id] = {
			id,
			title: `Project ${id}`,
			description: '',
		}
	}
}

describe('Home.vue last viewed section', () => {
	beforeEach(() => {
		mockAuthenticated.value = true
		mockShowLastViewed.value = true
		mockHasProjects.value = false
		mockGetHistory.mockReturnValue([{id: 1}, {id: 2}])

		for (const key of Object.keys(mockProjects)) {
			delete mockProjects[Number(key)]
		}
	})

	it('should show last viewed section when showLastViewed is true and history exists', async () => {
		seedProjects(1, 2)

		const wrapper = mountHome()
		await wrapper.vm.$nextTick()

		const heading = wrapper.find('h3')
		expect(heading.exists()).toBe(true)
		expect(heading.text()).toBe('Last viewed')
		expect(wrapper.find('.project-card-grid-stub').exists()).toBe(true)
	})

	it('should hide last viewed section when showLastViewed is false', async () => {
		seedProjects(1, 2)
		mockShowLastViewed.value = false

		const wrapper = mountHome()
		await wrapper.vm.$nextTick()

		expect(wrapper.find('h3').exists()).toBe(false)
		expect(wrapper.find('.project-card-grid-stub').exists()).toBe(false)
	})

	it('should hide last viewed section when project history is empty', async () => {
		mockGetHistory.mockReturnValue([])

		const wrapper = mountHome()
		await wrapper.vm.$nextTick()

		expect(wrapper.find('h3').exists()).toBe(false)
		expect(wrapper.find('.project-card-grid-stub').exists()).toBe(false)
	})

	it('should hide last viewed section when user is not authenticated', async () => {
		seedProjects(1, 2)
		mockAuthenticated.value = false

		const wrapper = mountHome()
		await wrapper.vm.$nextTick()

		expect(wrapper.find('h3').exists()).toBe(false)
		expect(wrapper.find('.project-card-grid-stub').exists()).toBe(false)
	})

	it('should show last viewed section when showLastViewed is undefined (backwards compat)', async () => {
		seedProjects(1, 2)
		mockShowLastViewed.value = undefined as any

		const wrapper = mountHome()
		await wrapper.vm.$nextTick()

		const heading = wrapper.find('h3')
		expect(heading.exists()).toBe(true)
		expect(heading.text()).toBe('Last viewed')
	})
})
