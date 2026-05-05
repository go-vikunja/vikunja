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
				:aria-label="ariaLabel"
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
	ariaLabel?: string,
}>(), {
	modelValue: false,
	ariaLabel: undefined,
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
	position: relative;
}

// Extend the hit target to >=44x44 without affecting layout (WCAG 2.5.5).
.base-checkbox__label::before {
	content: '';
	position: absolute;
	inset-block-start: 50%;
	inset-inline-start: 50%;
	min-block-size: 44px;
	min-inline-size: 44px;
	block-size: 100%;
	inline-size: 100%;
	transform: translate(-50%, -50%);
}

.base-checkbox:has(input:disabled) .base-checkbox__label {
	cursor: not-allowed;
	pointer-events: none;
}
</style>
