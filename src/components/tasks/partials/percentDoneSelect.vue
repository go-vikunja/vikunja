<template>
	<div class="select">
		<select :disabled="disabled || null" v-model.number="percentDone">
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

<script lang="ts">
import {defineComponent} from 'vue'
export default defineComponent({
	name: 'percentDoneSelect',
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
	computed: {
		percentDone: {
			get() {
				return this.modelValue
			},
			set(percentDone) {
				this.$emit('update:modelValue', percentDone)
				this.$emit('change')
			},
		},
	},
})
</script>
