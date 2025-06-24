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
import GanttChartPrimitive from '../primitives/GanttChartPrimitive.vue'
defineProps<{ rows:string[]; cellsByRow:Record<string,string[]> }>()
defineEmits<{(e:'update:focused', payload:{row:string|null;cell:number|null}):void}>()
</script>
