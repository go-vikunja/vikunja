<template>
	<GanttRowPrimitive
		:id="id"
		:selected="selected"
		@focus="onFocus"
		@select="onSelect"
	>
		<div
			class="w-full flex items-center px-3 py-2 border-b"
			:class="selected ? 'bg-row-selected' : (index % 2 ? 'bg-row-alt' : 'bg-row')"
			role="presentation"
		>
			<slot />
		</div>
	</GanttRowPrimitive>
</template>

<script setup lang="ts">
import GanttRowPrimitive from '../primitives/GanttRowPrimitive.vue'
const props = defineProps<{ id:string; index:number; selected:boolean }>()
const emit = defineEmits(['focus','select'])
const onFocus = () => emit('focus', props.id)
const onSelect = () => emit('select', props.id)
</script>

<style scoped>
.bg-row { background: var(--row-bg); }
.bg-row-alt { background: var(--row-alt-bg); }
.bg-row-selected { background: var(--row-selected-bg); }
.border-b { border-bottom:1px solid #ccc; }
</style>
