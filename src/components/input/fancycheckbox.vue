<template>
	<div :class="{'is-disabled': disabled}" class="fancycheckbox">
		<input
			:checked="checked"
			:disabled="disabled || null"
			:id="checkBoxId"
			@change="(event) => updateData(event.target.checked)"
			style="display: none;"
			type="checkbox"/>
		<label :for="checkBoxId" class="check">
			<svg height="18px" viewBox="0 0 18 18" width="18px">
				<path
					d="M1,9 L1,3.5 C1,2 2,1 3.5,1 L14.5,1 C16,1 17,2 17,3.5 L17,14.5 C17,16 16,17 14.5,17 L3.5,17 C2,17 1,16 1,14.5 L1,9 Z"></path>
				<polyline points="1 9 7 14 15 4"></polyline>
			</svg>
			<span>
				<slot></slot>
			</span>
		</label>
	</div>
</template>

<script>
import {createRandomID} from '@/helpers/randomId'

export default {
	name: 'fancycheckbox',
	data() {
		return {
			checked: false,
			checkBoxId: `fancycheckbox_${createRandomID()}`,
		}
	},
	props: {
		modelValue: {
			required: false,
		},
		disabled: {
			type: Boolean,
			required: false,
			default: false,
		},
	},
	emits: ['update:modelValue', 'change'],
	watch: {
		modelValue: {
			handler(modelValue) {
				this.checked = modelValue

			},
			immediate: true,
		},
	},
	methods: {
		updateData(checked) {
			this.checked = checked
			this.$emit('update:modelValue', checked)
			this.$emit('change', checked)
		},
	},
}
</script>
