export interface GanttPanePredecessor {
	id: number
	label: string
}

// One row of the task table next to the chart, in the order of the chart rows.
export interface GanttPaneRow {
	id: number
	wbs: string
	title: string
	indent: number
	isParent: boolean
	isMilestone: boolean
	isCritical: boolean
	isDone: boolean
	// The dates the task really has, null when it has none. The bar may show a stand-in.
	start: Date | null
	end: Date | null
	// Calendar days, null without both dates.
	duration: number | null
	// 0 to 100
	percent: number
	predecessors: GanttPanePredecessor[]
	assignees: string
	// The change of the baseline in days, positive means later than planned. Null without a baseline.
	variance: number | null
}

// The two header tiers of the chart (32 and 52 pixels) plus the border below them. Keep in sync
// with GanttTimelineHeader.vue.
export const GANTT_HEADER_HEIGHT_PX = 85
export const GANTT_ROW_HEIGHT_PX = 40
