<script setup lang="ts">
import {computed, useSlots, useId, ref} from 'vue'

interface Props {
	modelValue?: string | number
	label?: string
	error?: string | null
	id?: string
	disabled?: boolean
	loading?: boolean
	layout?: 'stacked' | 'two-col'
}

const props = withDefaults(defineProps<Props>(), {
	layout: 'stacked',
})
const emit = defineEmits<{
	'update:modelValue': [value: string | number]
}>()

function handleInput(event: Event) {
	const value = (event.target as HTMLInputElement).value
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

const inputBindings = computed(() => {
	const bindings: Record<string, unknown> = {}
	if (props.modelValue !== undefined) {
		bindings.value = props.modelValue
	}
	return bindings
})

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
		<template v-if="layout === 'two-col'">
			<label
				v-if="label"
				class="two-col"
			>
				<span>{{ label }}</span>
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
			</label>
			<div
				v-if="$slots.addon"
				class="control"
			>
				<slot name="addon" />
			</div>
		</template>
		<template v-else>
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
		</template>
		<p
			v-if="error"
			class="help is-danger"
		>
			{{ error }}
		</p>
	</div>
</template>

<style lang="scss" scoped>
label.two-col {
	display: flex;
	align-items: center;
	gap: .5rem;
}

label.two-col > span,
label.two-col :deep(input),
label.two-col :deep(.input),
label.two-col :deep(.select),
label.two-col :deep(.timezone-select),
label.two-col :deep(.multiselect) {
	flex: 0 0 50%;
	box-sizing: border-box;
}
</style>
