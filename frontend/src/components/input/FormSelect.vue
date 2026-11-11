<script setup lang="ts">
import {computed, useId} from 'vue'
import {useDelayedLoading} from '@/composables/useDelayedLoading'

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
	flat?: boolean
}

const props = withDefaults(defineProps<Props>(), {
	modelModifiers: () => ({}),
})
const emit = defineEmits<{
	'update:modelValue': [value: string | number]
}>()
const showLoading = useDelayedLoading(() => props.loading)

defineOptions({inheritAttrs: false})

const fallbackId = useId()
const selectId = computed(() => props.id ?? fallbackId)
const errorId = computed(() => props.error ? `${selectId.value}-error` : undefined)

const wrapperClasses = computed(() => [
	'select',
	{
		'is-loading': showLoading.value,
		'is-flat': props.flat,
	},
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
			:aria-invalid="error ? true : undefined"
			:aria-describedby="errorId"
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
		:id="errorId"
		class="help is-danger"
		role="alert"
	>
		{{ error }}
	</p>
</template>

<style lang="scss" scoped>
.select select {
	inline-size: 100%;
}

.select.is-flat {
	inline-size: 100%;
	// Bulma sizes the wrapper to --input-height.
	block-size: auto;
	margin-block: .1rem;

	select {
		background: transparent;
		border-color: transparent;
		box-shadow: none;
		block-size: 2rem;
		padding-block: 0;
		padding-inline-start: .5rem;
		line-height: 2rem;
	}

	&::after {
		opacity: 0;
	}

	&:hover::after {
		opacity: 1;
	}

	select:disabled {
		color: var(--text);
		opacity: 1;
		cursor: default;
	}

	&:has(select:disabled)::after {
		display: none;
	}

	select:not(:disabled) {
		cursor: pointer;

		&:hover,
		&:focus {
			background: var(--scheme-main);
		}

		&:focus-visible {
			border-color: var(--primary);
			@include focus-ring;
		}
	}
}
</style>
