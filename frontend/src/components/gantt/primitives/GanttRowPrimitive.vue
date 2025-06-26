<template>
	<div
		role="row"
		:aria-selected="selected"
		tabindex="-1"
		:data-state="selected ? 'selected' : ''"
		@click="onSelect"
		@focus="onFocus"
	>
		<slot :selected="selected" />
	</div>
</template>

<script setup lang="ts">
const props = defineProps<{ 
  id: string
  selected: boolean 
}>()

const emit = defineEmits<{
  select: [id: string]
  focus: [id: string]
}>()

function onSelect() { 
	emit('select', props.id) 
}

function onFocus() { 
	emit('focus', props.id) 
}
</script>
