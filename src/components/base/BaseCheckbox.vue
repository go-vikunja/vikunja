<template>
	<div class="base-checkbox" v-cy="'checkbox'">
		<input
			type="checkbox"
			:id="checkboxId"
			class="is-sr-only"
			:checked="modelValue"
			@change="(event) => emit('update:modelValue', (event.target as HTMLInputElement).checked)"
			:disabled="disabled || undefined"
		/>

		<slot name="label" :checkboxId="checkboxId">
			<label :for="checkboxId" class="base-checkbox__label">
				<slot/>
			</label>
		</slot>
	</div>
</template>

<script setup lang="ts">
import {ref} from 'vue'
import {createRandomID} from '@/helpers/randomId'

defineProps({
	modelValue: {
		type: Boolean,
		default: false,
	},
	disabled: {
		type: Boolean,
		default: false,
	},
})

const emit = defineEmits<{
  (event: 'update:modelValue', value: boolean): void
}>()

const checkboxId = ref(`fancycheckbox_${createRandomID()}`)
</script>

<style lang="scss" scoped>
.base-checkbox__label {
	cursor: pointer;
	user-select: none;
	-webkit-tap-highlight-color: transparent;
	display: inline-flex;
}

.base-checkbox:has(input:disabled) .base-checkbox__label {
	cursor:not-allowed;
	pointer-events: none;
}
</style>