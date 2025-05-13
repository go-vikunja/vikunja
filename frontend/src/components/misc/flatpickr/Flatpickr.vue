<template>
	<input
		v-bind="attrs"
		ref="root"
		type="text"
		data-input
		:disabled="disabled"
	>
</template>

<script lang="ts">
import flatpickr from 'flatpickr'
import 'flatpickr/dist/flatpickr.css'

// FIXME: Not sure how to alias these correctly
// import Options = Flatpickr.Options doesn't work
type Hook = flatpickr.Options.Hook
type HookKey = flatpickr.Options.HookKey
type Options = flatpickr.Options.Options
type DateOption = flatpickr.Options.DateOption

function camelToKebab(string: string) {
	return string.replace(/([a-z])([A-Z])/g, '$1-$2').toLowerCase()
}

function arrayify<T = unknown>(obj: T) {
	return obj instanceof Array
		? obj
		: [obj]
}

function nullify<T = unknown>(value: T) {
	return (value && (value as unknown[]).length)
		? value
		: null
}

// Events to emit, copied from flatpickr source
const includedEvents = [
	'onChange',
	'onClose',
	'onDestroy',
	'onMonthChange',
	'onOpen',
	'onYearChange',
] as HookKey[]

// Let's not emit these events by default
const excludedEvents = [
	'onValueUpdate',
	'onDayCreate',
	'onParseConfig',
	'onReady',
	'onPreCalendarPosition',
	'onKeyDown',
] as HookKey[]

// Keep a copy of all events for later use
const allEvents = includedEvents.concat(excludedEvents)

export default {inheritAttrs: false}
</script>

<script setup lang="ts">
import {computed, onBeforeUnmount, onMounted, ref, toRefs, useAttrs, watch, watchEffect} from 'vue'

const props = withDefaults(defineProps<{
	modelValue: DateOption | DateOption[] | null,
	/**  https://flatpickr.js.org/options/ */
	config?: Options,
	events?: HookKey[],
	disabled?: boolean,
}>(), {
	config: () => ({
		defaultDate: undefined,
		wrap: false,
	}),
	events: () => includedEvents,
	disabled: false,
})

const emit = defineEmits([
	'blur',
	'update:modelValue',
	...allEvents.map(camelToKebab),
])

const {modelValue, config, disabled} = toRefs(props)

// bind listener like onBlur
const attrs = useAttrs()

const root = ref<HTMLInputElement | null>(null)
const fp = ref<flatpickr.Instance | null>(null)
// eslint-disable-next-line vue/no-setup-props-reactivity-loss
const safeConfig = ref<Options>({...props.config})

function prepareConfig() {
	// Don't mutate original object on parent component
	const newConfig: Options = {...props.config}

	props.events.forEach((hook) => {
		// Respect global callbacks registered via setDefault() method
		const globalCallbacks = flatpickr.defaultConfig[hook] || []

		// Inject our own method along with user callback
		const localCallback: Hook = (...args) => emit(camelToKebab(hook), ...args)

		// Overwrite with merged array
		newConfig[hook] = arrayify(newConfig[hook] || []).concat(
			globalCallbacks,
			localCallback,
		)
	})

	// Watch for value changed by date-picker itself and notify parent component
	const onChange: Hook = (dates) => emit('update:modelValue', dates)
	newConfig['onChange'] = arrayify(newConfig['onChange'] || []).concat(onChange)

	// Flatpickr does not emit input event in some cases
	// const onClose: Hook = (_selectedDates, dateStr) => emit('update:modelValue', dateStr)
	// newConfig['onClose'] = arrayify(newConfig['onClose'] || []).concat(onClose)

	// Set initial date without emitting any event
	newConfig.defaultDate = props.modelValue || newConfig.defaultDate

	safeConfig.value = newConfig

	return safeConfig.value
}

onMounted(() => {
	if (
		fp.value || // Return early if flatpickr is already loaded
		!root.value // our input needs to be mounted
	) {
		return
	}

	prepareConfig()

	/**
	 * Get the HTML node where flatpickr to be attached
	 * Bind on parent element if wrap is true
	 */
	const element = props.config.wrap
		? root.value.parentNode
		: root.value

	// Init flatpickr
	fp.value = flatpickr(element, safeConfig.value)
})
onBeforeUnmount(() => fp.value?.destroy())

watch(config, () => {
	if (!fp.value) return
	// Workaround: Don't pass hooks to configs again otherwise
	// previously registered hooks will stop working
	// Notice: we are looping through all events
	// This also means that new callbacks can not be passed once component has been initialized
	allEvents.forEach((hook) => {
		delete safeConfig.value?.[hook]
	})
	fp.value.set(safeConfig.value)

	// Passing these properties in `set()` method will cause flatpickr to trigger some callbacks
	const configCallbacks = ['locale', 'showMonths'] as (keyof Options)[]

	// Workaround: Allow to change locale dynamically
	configCallbacks.forEach(name => {
		if (typeof safeConfig.value?.[name] !== 'undefined' && fp.value) {
			fp.value.set(name, safeConfig.value[name])
		}
	})
}, {deep: true})

const fpInput = computed(() => {
	if (!fp.value) return
	return fp.value.altInput || fp.value.input
})

/**
 * init blur event
 * (is required by many validation libraries)
 */
function onBlur(event: Event) {
	emit('blur', nullify((event.target as HTMLInputElement).value))
}

watchEffect(() => fpInput.value?.addEventListener('blur', onBlur))
onBeforeUnmount(() => fpInput.value?.removeEventListener('blur', onBlur))

/**
 * Watch for the disabled property and sets the value to the real input.
 */
watchEffect(() => {
	if (disabled.value) {
		fpInput.value?.setAttribute('disabled', '')
	} else {
		fpInput.value?.removeAttribute('disabled')
	}
})

/**
 * Watch for changes from parent component and update DOM
 */
watch(
	modelValue,
	newValue => {
		// Prevent updates if v-model value is same as input's current value
		if (!root.value || newValue === nullify(root.value.value)) return
		// Make sure we have a flatpickr instance and
		// notify flatpickr instance that there is a change in value
		fp.value?.setDate(newValue, true)
	},
	{deep: true},
)
</script>
