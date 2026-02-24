import {computed, type Ref} from 'vue'
import {useProjectStore} from '@/stores/projects'
import type {IProject} from '@/modelTypes/IProject'

// A palette of distinct, accessible colors for sub-project identification
// Index 0 is reserved for the parent project
const PROJECT_COLORS = [
	'#A0AEC0', // Gray (parent)
	'#4299E1', // Blue
	'#48BB78', // Green
	'#ED8936', // Orange
	'#9F7AEA', // Purple
	'#F56565', // Red
	'#38B2AC', // Teal
	'#ECC94B', // Yellow
	'#ED64A6', // Pink
	'#667EEA', // Indigo
	'#FC8181', // Light Red
	'#68D391', // Light Green
	'#63B3ED', // Light Blue
]

export interface SubprojectColorEntry {
	id: number
	title: string
	color: string
}

export function useSubprojectColors(parentProjectId: Ref<number>) {
	const projectStore = useProjectStore()

	const parentProject = computed(() => {
		if (!parentProjectId.value || parentProjectId.value <= 0) return null
		return projectStore.projects[parentProjectId.value] || null
	})

	const childProjects = computed(() => {
		if (!parentProjectId.value || parentProjectId.value <= 0) return []
		return Object.values(projectStore.projects)
			.filter(p => p.parentProjectId === parentProjectId.value)
			.sort((a, b) => a.title.localeCompare(b.title))
	})

	const colorMap = computed(() => {
		const map = new Map<number, string>()
		// Parent project gets index 0 color
		if (parentProjectId.value > 0) {
			map.set(parentProjectId.value, PROJECT_COLORS[0])
		}
		childProjects.value.forEach((project, index) => {
			map.set(project.id, PROJECT_COLORS[(index + 1) % PROJECT_COLORS.length])
		})
		return map
	})

	const legend = computed((): SubprojectColorEntry[] => {
		const entries: SubprojectColorEntry[] = []
		// Parent project first
		if (parentProject.value) {
			entries.push({
				id: parentProject.value.id,
				title: parentProject.value.title,
				color: PROJECT_COLORS[0],
			})
		}
		// Then children
		childProjects.value.forEach((project, index) => {
			entries.push({
				id: project.id,
				title: project.title,
				color: PROJECT_COLORS[(index + 1) % PROJECT_COLORS.length],
			})
		})
		return entries
	})

	function getProjectColor(projectId: number): string | null {
		return colorMap.value.get(projectId) || null
	}

	return {
		childProjects,
		colorMap,
		legend,
		getProjectColor,
	}
}
