<template>
	<slot
		name="trigger"
		:is-open="openValue"
		:toggle="toggle"
		:close="close"
	/>
	<div
		ref="popup"
		class="popup"
		:class="{
			'is-open': openValue,
			'has-overflow': hasOverflow && openValue
		}"
		:inert="!openValue"
		@focusin="rememberFocusEntered"
	>
		<slot
			name="content"
			:is-open="openValue"
			:toggle="toggle"
			:close="close"
		/>
	</div>
</template>

<script setup lang="ts">
import {ref, watch, watchEffect} from 'vue'
import {onClickOutside, onKeyStroke} from '@vueuse/core'

const props = withDefaults(defineProps<{
	hasOverflow?: boolean
	open?: boolean
	ignoreClickClasses?: string[]
}>(), {
	hasOverflow: false,
	open: false,
	ignoreClickClasses: () => [],
})

const emit = defineEmits<{
	'update:open': [open: boolean]
}>()

defineSlots<{
	trigger(props: {
		isOpen: boolean,
		toggle: () => boolean,
		close: () => void,
	}) : void
	content(props: {
		isOpen: boolean,
		toggle: () => boolean,
		close: () => void
	}): void
}>()

// eslint-disable-next-line vue/no-setup-props-reactivity-loss
const openValue = ref(props.open)
watchEffect(() => {
	openValue.value = props.open
})

function close() {
	if (!openValue.value) {
		return
	}
	openValue.value = false
	emit('update:open', false)
}

// onClickOutside listens in the capture phase, so a trigger's `@click.stop` cannot keep it from
// closing first — without this guard the trigger's own toggle() reopens what the same click closed.
let closedByClickOutside = false

function toggle() {
	if (closedByClickOutside) {
		closedByClickOutside = false
		return false
	}
	openValue.value = !openValue.value
	emit('update:open', openValue.value)
	return openValue.value
}

const popup = ref<HTMLElement | null>(null)

let lastFocused: HTMLElement | null = null
let focusEnteredPopup = false

function rememberFocusEntered() {
	focusEnteredPopup = true
}

// Pre-flush so focus is restored before `inert` blurs it to <body>; immediate so a popup mounted open still records its trigger.
watch(openValue, open => {
	if (open) {
		lastFocused = document.activeElement as HTMLElement | null
		return
	}

	// An outside click blurs to <body> at mousedown, before onClickOutside fires, so <body> still means nothing else took focus.
	const active = document.activeElement
	if (focusEnteredPopup && lastFocused?.isConnected && (popup.value?.contains(active) || active === document.body)) {
		lastFocused.focus()
	}
	lastFocused = null
	focusEnteredPopup = false
}, {immediate: true})

onClickOutside(popup, (event) => {
	const target = event.target as HTMLElement
	// Check if the click target has any of the ignored classes
	if (target?.classList && props.ignoreClickClasses.some(className => target.classList.contains(className))) {
		return
	}
	if (!openValue.value) {
		return
	}
	closedByClickOutside = true
	setTimeout(() => {
		closedByClickOutside = false
	})
	close()
})

onKeyStroke('Escape', event => {
	// defaultPrevented means an inner control (flatpickr, Multiselect, …) already consumed this Escape.
	if (!openValue.value || event.defaultPrevented) {
		return
	}

	// Scope to the popup owning focus — lastFocused is the trigger, which keeps focus after opening.
	const target = event.target as Node | null
	if (!target || (!popup.value?.contains(target) && target !== lastFocused)) {
		return
	}

	// Cancels the close request of a wrapping native <dialog> so only the popup closes.
	event.preventDefault()
	close()
})
</script>

<style scoped lang="scss">
.popup {
	transition: opacity $transition;
	opacity: 0;
	visibility: hidden;
	block-size: 0;
	overflow: hidden;
	position: absolute;
	inset-block-start: 1rem;
	z-index: 100;

	&.is-open {
		opacity: 1;
		visibility: visible;
		block-size: auto;
	}
}
</style>
