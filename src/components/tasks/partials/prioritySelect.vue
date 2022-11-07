<template>
	<div class="select">
		<select
			v-model="priority"
			@change="updateData"
			:disabled="disabled || undefined"
		>
			<option :value="PRIORITIES.UNSET">{{ $t('task.priority.unset') }}</option>
			<option :value="PRIORITIES.LOW">{{ $t('task.priority.low') }}</option>
			<option :value="PRIORITIES.MEDIUM">{{ $t('task.priority.medium') }}</option>
			<option :value="PRIORITIES.HIGH">{{ $t('task.priority.high') }}</option>
			<option :value="PRIORITIES.URGENT">{{ $t('task.priority.urgent') }}</option>
			<option :value="PRIORITIES.DO_NOW">{{ $t('task.priority.doNow') }}</option>
		</select>
	</div>
</template>

<script setup lang="ts">
import {ref, watch} from 'vue'
import {PRIORITIES} from '@/constants/priorities'

const props = defineProps({
	modelValue: {
		type: Number,
		default: 0,
	},
	disabled: {
		default: false,
	},
})
const emit = defineEmits(['update:modelValue'])

const priority = ref(0)

watch(
	() => props.modelValue,
	(value) => {
		priority.value = value
	},
	{immediate: true},
)

function updateData() {
	emit('update:modelValue', priority.value)
}
</script>
