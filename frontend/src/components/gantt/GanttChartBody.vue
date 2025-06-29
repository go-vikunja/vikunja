<template>
	<GanttChartPrimitive
		:rows="rows"
		:cells-by-row="cellsByRow"
		@update:focused="$emit('update:focused', $event)"
	>
		<slot />
	</GanttChartPrimitive>
</template>

<script setup lang="ts">
import GanttChartPrimitive from '@/components/gantt/primitives/GanttChartPrimitive.vue'

defineProps<{ 
	rows: string[]; 
	cellsByRow: Record<string, string[]> 
}>()
defineEmits<{
	'update:focused': [payload: { row: string | null; cell: number | null }],
}>()
</script>
