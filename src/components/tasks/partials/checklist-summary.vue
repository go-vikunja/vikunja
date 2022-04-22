<template>
	<span v-if="checklist.total > 0" class="checklist-summary">
		<svg width="12" height="12">
			<circle stroke-width="2" fill="transparent" cx="50%" cy="50%" r="5"></circle>
			<circle stroke-width="2" stroke-dasharray="31" :stroke-dashoffset="checklistCircleDone"
					stroke-linecap="round" fill="transparent" cx="50%" cy="50%" r="5"></circle>
		</svg>
		<span>
			{{ $t(checklist.total === checklist.checked ? 'task.checklistAllDone' : 'task.checklistTotal', checklist) }}
		</span>
	</span>
</template>

<script lang="ts">
import {defineComponent} from 'vue'

import {getChecklistStatistics} from '@/helpers/checklistFromText'

export default defineComponent({
	name: 'checklist-summary',
	props: {
		task: {
			required: true,
		},
	},
	computed: {
		checklist() {
			return getChecklistStatistics(this.task.description)
		},
		checklistCircleDone() {
			const r = 5
			const c = Math.PI * (r * 2)

			const progress = this.checklist.checked / this.checklist.total * 100

			return ((100 - progress) / 100) * c
		},
	},
})
</script>

<style scoped lang="scss">
.checklist-summary {
	color: var(--grey-500);
	display: inline-flex;
	align-items: center;
	padding-left: .5rem;
	font-size: .9rem;

	svg {
		transform: rotate(-90deg);
		transition: stroke-dashoffset 0.35s;
		margin-right: .25rem;

		circle {
			stroke: var(--grey-400);

			&:last-child {
				stroke: var(--primary);
			}
		}
	}
}
</style>