<template>
	<div class="select">
		<select :disabled="disabled" @change="updateData" v-model.number="percentDone">
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

<script>
export default {
	name: 'percentDoneSelect',
	data() {
		return {
			percentDone: 0,
		}
	},
	props: {
		value: {
			default: 0,
			type: Number,
		},
		disabled: {
			default: false,
		},
	},
	watch: {
		// Set the priority to the :value every time it changes from the outside
		value(newVal) {
			this.percentDone = newVal
		},
	},
	mounted() {
		this.percentDone = this.value
	},
	methods: {
		updateData() {
			this.$emit('input', this.percentDone)
			this.$emit('change')
		},
	},
}
</script>
