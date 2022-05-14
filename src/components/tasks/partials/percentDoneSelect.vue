<template>
	<div class="select">
		<select
			v-model.number="percentDone"
			:disabled="disabled || undefined"
		>
			<option value="0">0%</option>
			<option value="0.1">10%</option>
			<option value="0.2">20%</option>
			<option value="0.3">30%</option>
			<option value="0.4">40%</option>
			<option value="0.5">50%</option>
			<option value="0.6">60%</option>
			<option value="0.7">70%</option>
			<option value="0.8">80%</option>
			<option value="0.9">90%</option>
			<option value="1">100%</option>
		</select>
	</div>
</template>

<script setup lang="ts">
import {computed} from 'vue'

const props = defineProps({
	modelValue: {
		default: 0,
		type: Number,
	},
	disabled: {
		default: false,
	},
})

const emit = defineEmits(['update:modelValue', 'change'])

const percentDone = computed({
	get: () => props.modelValue,
	set(percentDone) {
		emit('update:modelValue', percentDone)
		emit('change')
	},
})
</script>
