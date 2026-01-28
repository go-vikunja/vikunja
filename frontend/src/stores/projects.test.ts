import {setActivePinia, createPinia} from 'pinia'
import {describe, it, expect, beforeEach, vi} from 'vitest'

import {useProjectStore} from './projects'

import type {IProject} from '@/modelTypes/IProject'

// Mock the dependencies that the store imports
vi.mock('vue-router', () => ({
	useRouter: () => ({
		push: vi.fn(),
	}),
}))

vi.mock('vue-i18n', () => ({
	useI18n: () => ({
		t: (key: string) => key,
	}),
	createI18n: () => ({
		global: {
			t: (key: string) => key,
		},
	}),
}))

vi.mock('@/stores/base', () => ({
	useBaseStore: () => ({
		currentProject: null,
		setCurrentProject: vi.fn(),
	}),
}))

vi.mock('@/indexes', () => ({
	createNewIndexer: () => ({
		add: vi.fn(),
		remove: vi.fn(),
		search: vi.fn(),
		update: vi.fn(),
	}),
}))

function createMockProject(overrides: Partial<IProject>): IProject {
	return {
		id: 1,
		title: 'Test Project',
		description: '',
		owner: {id: 1, username: 'test', name: '', email: '', created: new Date(), updated: new Date()},
		tasks: [],
		isArchived: false,
		hexColor: '',
		identifier: '',
		backgroundInformation: null,
		isFavorite: false,
		subscription: null as any,
		position: 0,
		backgroundBlurHash: '',
		parentProjectId: 0,
		views: [],
		created: new Date(),
		updated: new Date(),
		...overrides,
	} as IProject
}

describe('project store', () => {
	beforeEach(() => {
		setActivePinia(createPinia())
	})

	describe('notArchivedRootProjects', () => {
		it('should include root projects (parentProjectId === 0)', () => {
			const store = useProjectStore()
			const rootProject = createMockProject({id: 1, parentProjectId: 0, title: 'Root'})

			store.setProject(rootProject)

			expect(store.notArchivedRootProjects).toHaveLength(1)
			expect(store.notArchivedRootProjects[0].title).toBe('Root')
		})

		it('should exclude archived projects', () => {
			const store = useProjectStore()
			const archivedProject = createMockProject({id: 1, parentProjectId: 0, isArchived: true})

			store.setProject(archivedProject)

			expect(store.notArchivedRootProjects).toHaveLength(0)
		})

		it('should exclude saved filters (id < 0)', () => {
			const store = useProjectStore()
			const savedFilter = createMockProject({id: -2, parentProjectId: 0})

			store.setProject(savedFilter)

			expect(store.notArchivedRootProjects).toHaveLength(0)
		})

		it('should exclude sub-projects when parent is accessible', () => {
			const store = useProjectStore()
			const parentProject = createMockProject({id: 1, parentProjectId: 0, title: 'Parent'})
			const childProject = createMockProject({id: 2, parentProjectId: 1, title: 'Child'})

			store.setProject(parentProject)
			store.setProject(childProject)

			// Only parent should be in root projects
			expect(store.notArchivedRootProjects).toHaveLength(1)
			expect(store.notArchivedRootProjects[0].title).toBe('Parent')
		})

		it('should include orphaned sub-projects (parent not accessible)', () => {
			const store = useProjectStore()
			// Sub-project with parentProjectId pointing to a project not in the store
			const orphanedProject = createMockProject({id: 2, parentProjectId: 999, title: 'Orphaned'})

			store.setProject(orphanedProject)

			// Orphaned project should appear as a root project
			expect(store.notArchivedRootProjects).toHaveLength(1)
			expect(store.notArchivedRootProjects[0].title).toBe('Orphaned')
		})

		it('should handle mixed scenario with root, child, and orphaned projects', () => {
			const store = useProjectStore()
			const rootProject = createMockProject({id: 1, parentProjectId: 0, title: 'Root', position: 1})
			const childProject = createMockProject({id: 2, parentProjectId: 1, title: 'Child', position: 2})
			const orphanedProject = createMockProject({id: 3, parentProjectId: 999, title: 'Orphaned', position: 3})

			store.setProject(rootProject)
			store.setProject(childProject)
			store.setProject(orphanedProject)

			// Root and orphaned should be in root projects, but not child
			expect(store.notArchivedRootProjects).toHaveLength(2)
			const titles = store.notArchivedRootProjects.map(p => p.title)
			expect(titles).toContain('Root')
			expect(titles).toContain('Orphaned')
			expect(titles).not.toContain('Child')
		})
	})
})
