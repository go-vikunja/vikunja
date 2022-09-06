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

const priority = ref(0)

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

// FIXME: store value outside
// Set the priority to the :value every time it changes from the outside
watch(
	() => props.modelValue,
	(value) => {
		priority.value = value
	},
	{immediate: true},
)

function updateData() {
	emit('update:modelValue', priority.value)
	emit('change')
}
</script>
