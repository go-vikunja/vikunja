import {describe, it, expect, vi, beforeEach, afterEach} from 'vitest'
import {useRouteWithModal} from './useRouteWithModal'

// Mock vue-router
vi.mock('vue-router', () => ({
	useRouter: () => ({
		resolve: vi.fn((route) => ({resolved: route})),
		back: vi.fn(),
		push: vi.fn(),
	}),
	useRoute: () => ({
		fullPath: '/test',
		params: {},
		matched: [{}],
	}),
}))

// Mock stores
vi.mock('@/stores/base', () => ({
	useBaseStore: () => ({
		currentProject: null,
	}),
}))

vi.mock('@/stores/projects', () => ({
	useProjectStore: () => ({
		projects: {},
	}),
}))

describe('useRouteWithModal', () => {
	let originalHistoryState: any

	beforeEach(() => {
		// Save original history state
		originalHistoryState = window.history.state
	})

	afterEach(() => {
		// Restore original history state
		Object.defineProperty(window.history, 'state', {
			value: originalHistoryState,
			writable: true,
		})
	})

	it('should handle null history.state without throwing error', async () => {
		// Mock history.state as null (simulating direct navigation or reload)
		Object.defineProperty(window.history, 'state', {
			value: null,
			writable: true,
		})

		// This should not throw an error
		expect(() => {
			const {routeWithModal, currentModal} = useRouteWithModal()
			
			// Access computed values to trigger evaluation
			expect(routeWithModal.value).toBeDefined()
			expect(currentModal.value).toBeUndefined()
		}).not.toThrow()
	})

	it('should return route when no backdropView is present due to null state', async () => {
		// Mock history.state as null
		Object.defineProperty(window.history, 'state', {
			value: null,
			writable: true,
		})

		const {routeWithModal} = useRouteWithModal()
		
		// Should return the current route when history.state is null
		expect(routeWithModal.value).toEqual({
			fullPath: '/test',
			params: {},
			matched: [{}],
		})
	})
})