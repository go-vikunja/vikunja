<script setup lang="ts">
import {computed, ref, useId} from 'vue'

interface Props {
	modelValue?: string | number | Date | null
	modelModifiers?: {number?: boolean}
	id?: string
	disabled?: boolean
	loading?: boolean
	error?: string | null
}

const props = withDefaults(defineProps<Props>(), {
	modelModifiers: () => ({}),
})
const emit = defineEmits<{
	'update:modelValue': [value: string | number]
}>()


defineOptions({inheritAttrs: false})

const fallbackId = useId()
const inputId = computed(() => props.id ?? fallbackId)
const errorId = computed(() => props.error ? `${inputId.value}-error` : undefined)

const inputClasses = computed(() => [
	'input',
	{
		disabled: props.disabled,
		'is-loading': props.loading,
	},
])

const inputBindings = computed(() => {
	const bindings: Record<string, unknown> = {}
	if (props.modelValue !== undefined) {
		bindings.value = props.modelValue
	}
	return bindings
})

function handleInput(event: Event) {
	const value = (event.target as HTMLInputElement).value
	const shouldCoerceNumber = props.modelModifiers.number || typeof props.modelValue === 'number'
	if (shouldCoerceNumber) {
		emit('update:modelValue', value === '' ? '' : Number(value))
	} else {
		emit('update:modelValue', value)
	}
}

const inputRef = ref<HTMLInputElement | null>(null)
defineExpose({
	get value() {
		return inputRef.value?.value ?? ''
	},
	focus() {
		inputRef.value?.focus()
	},
})
</script>

<template>
	<input
		:id="inputId"
		ref="inputRef"
		v-bind="{ ...$attrs, ...inputBindings }"
		:class="inputClasses"
		:disabled="disabled || undefined"
		:aria-invalid="error ? true : undefined"
		:aria-describedby="errorId"
		@input="handleInput"
	>
	<p
		v-if="error"
		:id="errorId"
		class="help is-danger"
		role="alert"
	>
		{{ error }}
	</p>
</template>
