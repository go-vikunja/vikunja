import {getHexColor} from '@/models/task'

// The bar/chip colour for a task: its project's colour, falling back to the
// task's own and then the theme primary.
export function plannerTaskColor(taskHexColor: string, projectHexColor?: string): string {
	return getHexColor(projectHexColor ?? '') ?? getHexColor(taskHexColor) ?? 'var(--primary)'
}
