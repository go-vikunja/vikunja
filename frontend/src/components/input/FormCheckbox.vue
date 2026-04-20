<script setup lang="ts">
interface Props {
	modelValue?: boolean
	label?: string
	disabled?: boolean
}

defineProps<Props>()
const emit = defineEmits<{
	'update:modelValue': [value: boolean]
}>()

function handleChange(event: Event) {
	emit('update:modelValue', (event.target as HTMLInputElement).checked)
}
</script>

<template>
	<label class="checkbox">
		<input
			type="checkbox"
			:checked="modelValue"
			:disabled="disabled || undefined"
			@change="handleChange"
		>
		<slot>{{ label }}</slot>
	</label>
</template>

<style lang="scss" scoped>
// Ported from bulma-css-variables/sass/form/checkbox-radio.sass
// (the %checkbox-radio placeholder, scoped to .checkbox since this
// component is the sole consumer of that class).
label.checkbox {
	cursor: pointer;
	line-height: 1.25;
	position: relative;

	display: flex;
	align-items: center;
	gap: .5rem;
	inline-size: fit-content;

	&:hover {
		color: var(--input-hover-color);
	}

	&[disabled],
	input[disabled] {
		color: var(--input-disabled-color);
		cursor: not-allowed;
	}

	input {
		cursor: pointer;
	}

	&:not(:last-child) {
		margin-block-end: .75rem;
	}
}
</style>
