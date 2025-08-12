<template>
	<div
		role="row" 
		:aria-selected="selected" 
		tabindex="-1" 
		:data-state="selected ? 'selected' : null" 
		@click="onSelect"
		@focus="onFocus"
		@keydown="onKeyDown"
	>
		<slot :selected="selected" />
	</div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const props = defineProps<{
	id: string
}>()

const emit = defineEmits<{
	select: [id: string]
	focus: [id: string]
}>()

const selected = ref(false)

function onSelect() {
	emit('select', props.id)
}

function onFocus() {
	emit('focus', props.id)
}

function onKeyDown(e: KeyboardEvent) {
	if (e.key === 'Enter' || e.key === ' ') {
		e.preventDefault()
		onSelect()
	}
}
</script>
