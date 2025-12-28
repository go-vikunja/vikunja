import {ref, computed} from 'vue'
import TaskCollectionService from '@/services/taskCollection'
import type {IProject} from '@/modelTypes/IProject'

// Global cache for task counts - shared across all components
const taskCounts = ref<Record<number, number>>({})
const loadingProjects = ref<Set<number>>(new Set())
const fetchedProjects = ref<Set<number>>(new Set())

// Fetch task count for a single project using minimal API call
async function fetchTaskCount(projectId: number): Promise<number> {
	// Skip pseudo-projects (favorites=-1, inbox=0)
	if (projectId === 0 || projectId === -1) {
		return 0
	}

	// Already loading this project
	if (loadingProjects.value.has(projectId)) {
		return taskCounts.value[projectId] ?? 0
	}

	// Already fetched
	if (fetchedProjects.value.has(projectId)) {
		return taskCounts.value[projectId] ?? 0
	}

	loadingProjects.value.add(projectId)

	try {
		const service = new TaskCollectionService()
		// Make minimal request with per_page=1
		// totalPages header will equal total items when per_page=1
		await service.getAll({projectId}, {
			per_page: 1,
			filter: 'done = false',
			filter_include_nulls: false,
			sort_by: ['id'],
			order_by: ['asc'],
			s: '',
		})

		// With per_page=1, totalPages equals total number of items
		const count = service.totalPages || 0

		taskCounts.value[projectId] = count
		fetchedProjects.value.add(projectId)

		return count
	} catch (e) {
		console.error(`Failed to fetch task count for project ${projectId}:`, e)
		taskCounts.value[projectId] = 0
		fetchedProjects.value.add(projectId)
		return 0
	} finally {
		loadingProjects.value.delete(projectId)
	}
}

// Batch fetch task counts for multiple projects
async function fetchTaskCountsForProjects(projects: IProject[]): Promise<void> {
	const projectsToFetch = projects.filter(p =>
		(p.id > 0 || p.id < -1) &&
		!fetchedProjects.value.has(p.id) &&
		!loadingProjects.value.has(p.id),
	)

	// Fetch in parallel but limit concurrency to avoid overwhelming the API
	const batchSize = 5
	for (let i = 0; i < projectsToFetch.length; i += batchSize) {
		const batch = projectsToFetch.slice(i, i + batchSize)
		await Promise.all(batch.map(p => fetchTaskCount(p.id)))
	}
}

// Clear cache (useful when tasks are created/deleted)
function clearCache() {
	taskCounts.value = {}
	fetchedProjects.value.clear()
}

// Invalidate a specific project's cache
function invalidateProject(projectId: number) {
	delete taskCounts.value[projectId]
	fetchedProjects.value.delete(projectId)
}

export function useProjectTaskCounts() {
	const getTaskCount = computed(() => {
		return (projectId: number) => taskCounts.value[projectId] ?? null
	})

	const isLoading = computed(() => {
		return (projectId: number) => loadingProjects.value.has(projectId)
	})

	return {
		taskCounts,
		getTaskCount,
		isLoading,
		fetchTaskCount,
		fetchTaskCountsForProjects,
		clearCache,
		invalidateProject,
	}
}
