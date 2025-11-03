<template>
	<GanttRowPrimitive
		:id="id"
		@focus="onFocus"
		@select="onSelect"
	>
		<div
			class="w-full flex items-center"
			:class="index % 2 ? 'bg-row-alt' : 'bg-row'"
			role="presentation"
		>
			<slot />
		</div>
	</GanttRowPrimitive>
</template>

<script setup lang="ts">
import GanttRowPrimitive from '@/components/gantt/primitives/GanttRowPrimitive.vue'

const props = defineProps<{ 
	id: string
	index: number 
}>()
const emit = defineEmits<{
  (e: 'focus', id: string): void
  (e: 'select', id: string): void
}>()
const onFocus = () => emit('focus', props.id)
const onSelect = () => emit('select', props.id)
</script>

<style scoped>
.bg-row {
	background: hsla(var(--white-h), var(--white-s), var(--white-l), .15); 
}

.bg-row-alt {
	background: hsla(var(--grey-100-hsl), .5); 
}
</style>
