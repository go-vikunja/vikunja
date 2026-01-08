<script setup lang="ts">
import {computed, useSlots, useId, ref} from 'vue'

interface Props {
	modelValue?: string | number
	label?: string
	error?: string | null
	id?: string
}

const props = defineProps<Props>()
defineEmits<{
	'update:modelValue': [value: string]
}>()

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

// Expose input element for direct access (needed for browser autofill workarounds)
const inputRef = ref<HTMLInputElement | null>(null)
defineExpose({
	get value() {
		return inputRef.value?.value ?? ''
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
			<slot>
				<input
					:id="inputId"
					ref="inputRef"
					v-bind="$attrs"
					:value="modelValue"
					class="input"
					@input="$emit('update:modelValue', ($event.target as HTMLInputElement).value)"
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
