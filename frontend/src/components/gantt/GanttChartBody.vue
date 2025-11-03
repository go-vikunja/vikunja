<template>
	<GanttChartPrimitive
		ref="primitiveRef"
		:rows="rows"
		:cells-by-row="cellsByRow"
		@update:focused="$emit('update:focused', $event)"
		@enterPressed="$emit('enterPressed', $event)"
	>
		<template #default="{ focusedRow, focusedCell }">
			<slot
				:focused-row="focusedRow"
				:focused-cell="focusedCell"
			/>
		</template>
	</GanttChartPrimitive>
</template>

<script setup lang="ts">
import {ref} from 'vue'

import GanttChartPrimitive from '@/components/gantt/primitives/GanttChartPrimitive.vue'

defineProps<{
	rows: string[]
	cellsByRow: Record<string, string[]>
}>()
defineEmits<{
	'update:focused': [payload: { row: string | null; cell: number | null }],
	'enterPressed': [payload: { row: string; cell: number }],
}>()

const primitiveRef = ref<InstanceType<typeof GanttChartPrimitive> | null>(null)

function setFocus(rowId: string, cellIndex: number = 0) {
	primitiveRef.value?.setFocus(rowId, cellIndex)
}

defineExpose({
	setFocus,
})
</script>
