<template>
	<div class="select">
		<select :disabled="disabled || null" @change="updateData" v-model="priority">
			<option :value="priorities.UNSET">{{ $t('task.priority.unset') }}</option>
			<option :value="priorities.LOW">{{ $t('task.priority.low') }}</option>
			<option :value="priorities.MEDIUM">{{ $t('task.priority.medium') }}</option>
			<option :value="priorities.HIGH">{{ $t('task.priority.high') }}</option>
			<option :value="priorities.URGENT">{{ $t('task.priority.urgent') }}</option>
			<option :value="priorities.DO_NOW">{{ $t('task.priority.doNow') }}</option>
		</select>
	</div>
</template>

<script>
import priorites from '../../../models/constants/priorities'

export default {
	name: 'prioritySelect',
	data() {
		return {
			priorities: priorites,
			priority: 0,
		}
	},
	props: {
		modelValue: {
			default: 0,
			type: Number,
		},
		disabled: {
			default: false,
		},
	},
	emits: ['update:modelValue', 'change'],
	watch: {
		// Set the priority to the :value every time it changes from the outside
		modelValue: {
			handler(value) {
				this.priority = value
			},
			immediate: true,
		},
	},
	methods: {
		updateData() {
			this.$emit('update:modelValue', this.priority)
			this.$emit('change')
		},
	},
}
</script>
