<template>
	<div
		v-cy="'checkbox'"
		class="base-checkbox"
	>
		<label
			class="base-checkbox__label"
		>
			<input
				type="checkbox"
				class="is-sr-only"
				:checked="modelValue"
				:disabled="disabled || undefined"
				@change="(event) => emit('update:modelValue', (event.target as HTMLInputElement).checked)"
			>
			<slot />
		</label>
	</div>
</template>

<script setup lang="ts">
withDefaults(defineProps<{
	modelValue?: boolean,
	disabled: boolean,
}>(), {
	modelValue: false,
})

const emit = defineEmits<{
	(event: 'update:modelValue', value: boolean): void
}>()
</script>

<style lang="scss" scoped>
.base-checkbox__label {
	cursor: pointer;
	user-select: none;
	-webkit-tap-highlight-color: transparent;
	display: inline-flex;
}

.base-checkbox:has(input:disabled) .base-checkbox__label {
	cursor: not-allowed;
	pointer-events: none;
}
</style>
