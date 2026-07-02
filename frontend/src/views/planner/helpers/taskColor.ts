// The bar/chip colour for a task: its project's colour, falling back to the
// task's own and then the theme primary. Model constructors already normalise
// hexColor to a leading '#', but guard anyway for un-modelled inputs.
export function plannerTaskColor(taskHexColor: string, projectHexColor?: string): string {
	const hex = projectHexColor || taskHexColor
	if (!hex) {
		return 'var(--primary)'
	}
	return hex.startsWith('#') ? hex : `#${hex}`
}
