<template>
	<span
		v-if="checklist.total > 0"
		class="checklist-summary"
	>
		<svg
			width="12"
			height="12"
		>
			<circle
				stroke-width="2"
				fill="transparent"
				cx="50%"
				cy="50%"
				r="5"
			/>
			<circle
				stroke-width="2"
				stroke-dasharray="31"
				:stroke-dashoffset="checklistCircleDone"
				stroke-linecap="round"
				fill="transparent"
				cx="50%"
				cy="50%"
				r="5"
			/>
		</svg>
		<span>{{ label }}</span>
	</span>
</template>

<script setup lang="ts">
import {computed} from 'vue'
import { useI18n } from 'vue-i18n'

import {getChecklistStatistics} from '@/helpers/checklistFromText'
import type {ITask} from '@/modelTypes/ITask'

const props = defineProps<{
	task: ITask
}>()

const checklist = computed(() => getChecklistStatistics(props.task.description))

const checklistCircleDone = computed(() => {
	const r = 5
	const c = Math.PI * (r * 2)

	const progress = checklist.value.checked / checklist.value.total * 100

	return ((100 - progress) / 100) * c
})

const {t} = useI18n({useScope: 'global'})
const label = computed(() => {
	return checklist.value.total === checklist.value.checked 
		? t('task.checklistAllDone', checklist.value)
		: t('task.checklistTotal', checklist.value)
})
</script>

<style scoped lang="scss">
.checklist-summary {
	color: var(--grey-500);
	display: inline-flex;
	align-items: center;
	padding-inline-start: .5rem;
	font-size: .9rem;
}

svg {
	transform: rotate(-90deg);
	transition: stroke-dashoffset 0.35s;
	margin-inline-end: .25rem;
}

circle {
	stroke: var(--grey-400);

	&:last-child {
		stroke: var(--primary);
	}
}
</style>
