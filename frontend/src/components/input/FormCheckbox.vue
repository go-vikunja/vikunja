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
label.checkbox {
	display: flex;
	align-items: center;
	gap: .5rem;
	inline-size: fit-content;

	&:not(:last-child) {
		margin-block-end: .75rem;
	}
}
</style>
