import type {Label} from '@/client/generated'

// Most labels in one diagram. Measured with random projects: up to twelve labels that mostly
// stand apart, every overlap still gets a spot and the smallest circle stays readable.
export const MAX_VENN_SETS = 12

type LabeledTask = {labels?: Label[] | null}

// A region is identified by a bitmask over the selected label ids: bit i is set when the task
// carries selectedLabelIds[i]. Mask 0 holds the tasks that carry none of the selected labels.
export type VennMask = number

export function labelsInUse(tasks: LabeledTask[]): Label[] {
	const counts = new Map<number, {label: Label, count: number}>()
	for (const task of tasks) {
		for (const label of uniqueLabels(task)) {
			const entry = counts.get(label.id!)
			if (entry) {
				entry.count++
			} else {
				counts.set(label.id!, {label, count: 1})
			}
		}
	}
	return [...counts.values()]
		.sort((a, b) => b.count - a.count || (a.label.title ?? '').localeCompare(b.label.title ?? ''))
		.map(({label}) => label)
}

export function vennMaskOf(task: LabeledTask, selectedLabelIds: number[]): VennMask {
	const ids = new Set(uniqueLabels(task).map(label => label.id))
	return selectedLabelIds.reduce((mask, id, i) => ids.has(id) ? mask | (1 << i) : mask, 0)
}

// Only regions that hold at least one task are returned.
export function groupTasksByVennRegion<T extends LabeledTask>(tasks: T[], labelIds: number[]): Map<VennMask, T[]> {
	const regions = new Map<VennMask, T[]>()
	for (const task of tasks) {
		const mask = vennMaskOf(task, labelIds)
		regions.set(mask, [...(regions.get(mask) ?? []), task])
	}
	return regions
}

// Bits of the labels that are not hidden.
export function visibleMaskOf(labelIds: number[], hiddenLabelIds: number[]): VennMask {
	const hidden = new Set(hiddenLabelIds)
	return labelIds.reduce((mask, id, i) => hidden.has(id) ? mask : mask | (1 << i), 0)
}

function uniqueLabels(task: LabeledTask): Label[] {
	return (task.labels ?? []).filter((label, i, all) =>
		label.id !== undefined && all.findIndex(l => l.id === label.id) === i)
}
