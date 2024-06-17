<template>
	<div
		v-cy="'checkbox'"
		class="base-checkbox"
	>
		<input
			:id="checkboxId"
			type="checkbox"
			class="is-sr-only"
			:checked="modelValue"
			:disabled="disabled || undefined"
			@change="(event) => emit('update:modelValue', (event.target as HTMLInputElement).checked)"
		>

		<slot
			name="label"
			:checkbox-id="checkboxId"
		>
			<label
				:for="checkboxId"
				class="base-checkbox__label"
			>
				<slot />
			</label>
		</slot>
	</div>
</template>

<script setup lang="ts">
import {ref} from 'vue'
import {createRandomID} from '@/helpers/randomId'

withDefaults(defineProps<{
	modelValue?: boolean,
	disabled: boolean,
}>(), {
	modelValue: false,
})

const emit = defineEmits<{
  (event: 'update:modelValue', value: boolean): void
}>()

const checkboxId = ref(`checkbox_${createRandomID()}`)
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