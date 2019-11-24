<template>
	<div class="select">
		<select v-model="priority" @change="updateData">
			<option :value="priorities.UNSET">Unset</option>
			<option :value="priorities.LOW">Low</option>
			<option :value="priorities.MEDIUM">Medium</option>
			<option :value="priorities.HIGH">High</option>
			<option :value="priorities.URGENT">Urgent</option>
			<option :value="priorities.DO_NOW">DO NOW</option>
		</select>
	</div>
</template>

<script>
	import priorites from '../../../models/priorities'

	export default {
		name: 'prioritySelect',
		data() {
			return {
				priorities: priorites,
				priority: 0,
			}
		},
		props: {
			value: {
				default: 0,
				type: Number,
			}
		},
		watch: {
			// Set the priority to the :value every time it changes from the outside
			value(newVal) {
				this.priority = newVal
			},
		},
		mounted() {
			this.priority = this.value
		},
		methods: {
			updateData() {
				this.$emit('input', this.priority)
				this.$emit('change')
			}
		},
	}
</script>
