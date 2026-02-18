<script setup lang="ts">
import {computed, useSlots, useId, ref} from 'vue'

interface Props {
	modelValue?: string | number
	label?: string
	error?: string | null
	id?: string
	disabled?: boolean
	loading?: boolean
}

const props = defineProps<Props>()
const emit = defineEmits<{
	'update:modelValue': [value: string | number]
}>()

function handleInput(event: Event) {
	const value = (event.target as HTMLInputElement).value
	// Preserve numeric type if modelValue was a number
	if (typeof props.modelValue === 'number') {
		emit('update:modelValue', value === '' ? '' : Number(value))
	} else {
		emit('update:modelValue', value)
	}
}

defineOptions({
	inheritAttrs: false,
})

const slots = useSlots()
const generatedId = useId()

const inputId = computed(() => props.id ?? generatedId)
const hasAddon = computed(() => !!slots.addon)

const fieldClasses = computed(() => [
	'field',
	{'has-addons': hasAddon.value},
])

const controlClasses = computed(() => [
	'control',
	{'is-expanded': hasAddon.value},
])

const inputClasses = computed(() => [
	'input',
	{
		'disabled': props.disabled,
		'is-loading': props.loading,
	},
])

// Only bind value when modelValue is explicitly provided (not undefined)
// This allows the component to be used without v-model for native input behavior
const inputBindings = computed(() => {
	const bindings: Record<string, unknown> = {}
	if (props.modelValue !== undefined) {
		bindings.value = props.modelValue
	}
	return bindings
})

// Expose input element for direct access (needed for browser autofill workarounds)
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
	<div :class="fieldClasses">
		<label
			v-if="label"
			:for="inputId"
			class="label"
		>
			{{ label }}
		</label>

		<div :class="controlClasses">
			<slot :id="inputId">
				<input
					:id="inputId"
					ref="inputRef"
					v-bind="{ ...$attrs, ...inputBindings }"
					:class="inputClasses"
					:disabled="disabled || undefined"
					@input="handleInput"
				>
			</slot>
		</div>

		<div
			v-if="$slots.addon"
			class="control"
		>
			<slot name="addon" />
		</div>

		<p
			v-if="error"
			class="help is-danger"
		>
			{{ error }}
		</p>
	</div>
</template>
