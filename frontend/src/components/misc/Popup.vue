<template>
	<slot
		name="trigger"
		:is-open="openValue"
		:toggle="toggle"
		:close="close"
	/>
	<Modal
		v-if="asSheet && openValue"
		variant="sheet"
		:title="sheetTitle"
		@close="close"
	>
		<template
			v-if="$slots['header-action']"
			#header-action
		>
			<slot name="header-action" />
		</template>
		<slot
			name="content"
			:is-open="openValue"
			:toggle="toggle"
			:close="close"
		/>
	</Modal>
	<div
		v-else-if="!asSheet"
		ref="popup"
		class="popup"
		:class="{
			'is-open': openValue,
			'has-overflow': hasOverflow && openValue,
		}"
		:style="floatingStyle"
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
import {computed, onScopeDispose, ref, watch, watchEffect} from 'vue'
import {onClickOutside, onKeyStroke} from '@vueuse/core'
import {autoUpdate, computePosition, flip, offset, shift, type Placement} from '@floating-ui/dom'

import Modal from '@/components/misc/Modal.vue'
import {useIsMobile} from '@/composables/useIsMobile'

const props = withDefaults(defineProps<{
	hasOverflow?: boolean
	open?: boolean
	ignoreClickClasses?: string[]
	// Anchors the popup to `anchor` with floating-ui (flips and shifts to stay on screen). Without it, consumers position via CSS.
	placement?: Placement
	anchor?: HTMLElement | null
	sheetOnMobile?: boolean
	sheetTitle?: string
}>(), {
	hasOverflow: false,
	open: false,
	ignoreClickClasses: () => [],
	placement: undefined,
	anchor: null,
	sheetOnMobile: false,
	sheetTitle: '',
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
	'header-action'(): void
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

const isMobile = useIsMobile()
const asSheet = computed(() => props.sheetOnMobile && isMobile.value)

const floatingStyle = ref<Record<string, string>>({})
// 4rem app header ($navbar-height) plus a small margin.
const VIEWPORT_PADDING = {top: 72, right: 8, bottom: 8, left: 8}
let scrolledIntoView = false

async function updatePosition() {
	if (!props.anchor || !popup.value || !props.placement) {
		return
	}
	const {x, y} = await computePosition(props.anchor, popup.value, {
		placement: props.placement,
		strategy: 'absolute',
		// Top padding keeps a flipped popup out from under the fixed app header. When neither side fits
		// (short window, tall popup) stay on the requested side and scroll it into view instead of
		// letting bestFit push it above the viewport.
		middleware: [
			offset(4),
			flip({padding: VIEWPORT_PADDING, fallbackStrategy: 'initialPlacement'}),
			shift({padding: VIEWPORT_PADDING}),
		],
	})
	floatingStyle.value = {left: `${x}px`, top: `${y}px`}
	if (!scrolledIntoView) {
		scrolledIntoView = true
		popup.value.scrollIntoView({block: 'nearest', inline: 'nearest'})
	}
}

let stopAutoUpdate: (() => void) | null = null
watch([openValue, asSheet, () => props.anchor], ([open, sheet, anchor]) => {
	stopAutoUpdate?.()
	stopAutoUpdate = null
	scrolledIntoView = false
	if (!open || sheet || !props.placement || !anchor || !popup.value) {
		floatingStyle.value = {}
		return
	}
	stopAutoUpdate = autoUpdate(anchor, popup.value, updatePosition)
}, {flush: 'post'})

onScopeDispose(() => {
	stopAutoUpdate?.()
	stopAutoUpdate = null
})

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
	// defaultPrevented means an inner control (Multiselect, …) already consumed this Escape.
	if (asSheet.value || !openValue.value || event.defaultPrevented) {
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
