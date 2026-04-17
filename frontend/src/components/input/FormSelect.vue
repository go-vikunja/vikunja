<script setup lang="ts">
import {computed, useId} from 'vue'

export type SelectOption =
	| string
	| number
	| {value: string | number, label: string, disabled?: boolean}

interface Props {
	modelValue?: string | number | null
	modelModifiers?: {number?: boolean}
	id?: string
	disabled?: boolean
	loading?: boolean
	error?: string | null
	options?: SelectOption[]
}

const props = withDefaults(defineProps<Props>(), {
	modelModifiers: () => ({}),
})
const emit = defineEmits<{
	'update:modelValue': [value: string | number]
}>()

defineOptions({inheritAttrs: false})

const fallbackId = useId()
const selectId = computed(() => props.id ?? fallbackId)

const wrapperClasses = computed(() => [
	'select',
	{'is-loading': props.loading},
])

const selectBindings = computed(() => {
	const bindings: Record<string, unknown> = {}
	if (props.modelValue !== undefined) {
		bindings.value = props.modelValue
	}
	return bindings
})

const normalizedOptions = computed(() => {
	if (!props.options) {
		return null
	}
	return props.options.map(opt => {
		if (typeof opt === 'object' && opt !== null) {
			return opt
		}
		return {value: opt, label: String(opt)}
	})
})

function handleChange(event: Event) {
	const value = (event.target as HTMLSelectElement).value
	const shouldCoerceNumber = props.modelModifiers.number || typeof props.modelValue === 'number'
	if (shouldCoerceNumber) {
		emit('update:modelValue', value === '' ? '' : Number(value))
	} else {
		emit('update:modelValue', value)
	}
}
</script>

<template>
	<div :class="wrapperClasses">
		<select
			:id="selectId"
			v-bind="{ ...$attrs, ...selectBindings }"
			:disabled="disabled || undefined"
			@change="handleChange"
		>
			<template v-if="normalizedOptions">
				<option
					v-for="opt in normalizedOptions"
					:key="opt.value"
					:value="opt.value"
					:disabled="opt.disabled || undefined"
				>
					{{ opt.label }}
				</option>
			</template>
			<slot v-else />
		</select>
	</div>
	<p
		v-if="error"
		class="help is-danger"
	>
		{{ error }}
	</p>
</template>

<style lang="scss" scoped>
.select select {
	inline-size: 100%;
}
</style>
