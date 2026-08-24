import {getHexColor} from '@/models/task'

// The bar/chip colour for a task: its own colour first (matching every other
// view), falling back to its project's and then the theme primary.
export function plannerTaskColor(taskHexColor: string, projectHexColor?: string): string {
	return getHexColor(taskHexColor) ?? getHexColor(projectHexColor ?? '') ?? 'var(--primary)'
}
