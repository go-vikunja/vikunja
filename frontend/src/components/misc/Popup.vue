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
		popover="auto"
		class="popup"
		:class="{
			'is-open': openValue,
			'has-overflow': hasOverflow && openValue,
		}"
		:style="floatingStyle"
		@toggle="onToggle"
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
import {autoUpdate, computePosition, flip, offset, shift, size, type Placement} from '@floating-ui/dom'

import Modal from '@/components/misc/Modal.vue'
import {useIsMobile} from '@/composables/useIsMobile'

const props = withDefaults(defineProps<{
	hasOverflow?: boolean
	open?: boolean
	// Anchors the popup to `anchor` with floating-ui (flips and shifts to stay on screen).
	placement?: Placement
	anchor?: HTMLElement | null
	sheetOnMobile?: boolean
	sheetTitle?: string
}>(), {
	hasOverflow: false,
	open: false,
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

// A light dismiss closes the popup on pointerup, so a trigger click that is still to come would
// reopen what the same interaction closed.
let closedByLightDismiss = false

function toggle() {
	if (closedByLightDismiss) {
		closedByLightDismiss = false
		return false
	}
	openValue.value = !openValue.value
	emit('update:open', openValue.value)
	return openValue.value
}

const popup = ref<HTMLElement | null>(null)

const isMobile = useIsMobile()
const asSheet = computed(() => props.sheetOnMobile && isMobile.value)

let popoverShown = false

function onToggle(event: Event) {
	const newState = (event as ToggleEvent).newState
	popoverShown = newState === 'open'

	if (newState === 'open' || !openValue.value) {
		return
	}

	// Outside click or Escape: the browser closed the popover, so mirror it into our state.
	closedByLightDismiss = true
	setTimeout(() => {
		closedByLightDismiss = false
	})
	close()
}

// An effect, not a watch on openValue: a light dismiss followed by a reopen in the same tick leaves
// openValue unchanged, and only reconciling against the popover's real state shows it again.
watchEffect(() => {
	const el = popup.value
	if (!el) {
		// Switching to the mobile sheet unmounts the box; a remounted one is never showing.
		popoverShown = false
		return
	}

	if (openValue.value && !asSheet.value) {
		if (!popoverShown) {
			el.showPopover()
			popoverShown = true
		}
		return
	}

	if (popoverShown) {
		el.hidePopover()
		popoverShown = false
	}
}, {flush: 'post'})

const floatingStyle = ref<Record<string, string>>({})
// 4rem app header ($navbar-height) plus a small margin.
const VIEWPORT_PADDING = {top: 72, right: 8, bottom: 8, left: 8}

async function updatePosition() {
	if (!props.anchor || !popup.value || !props.placement) {
		return
	}
	let availableBlockSize = 0
	const {x, y} = await computePosition(props.anchor, popup.value, {
		placement: props.placement,
		strategy: 'fixed',
		// Top padding keeps a flipped popup out from under the fixed app header. When neither side fits
		// (short window, tall popup) stay on the requested side and let size() cap the popup to what is
		// left, so it scrolls inside itself instead of overflowing the viewport.
		middleware: [
			offset(4),
			flip({padding: VIEWPORT_PADDING, fallbackStrategy: 'initialPlacement'}),
			shift({padding: VIEWPORT_PADDING}),
			size({
				padding: VIEWPORT_PADDING,
				apply: ({availableHeight}) => {
					availableBlockSize = availableHeight
				},
			}),
		],
	})
	floatingStyle.value = {
		left: `${x}px`,
		top: `${y}px`,
		maxBlockSize: `${Math.max(availableBlockSize, 0)}px`,
	}
}

let stopAutoUpdate: (() => void) | null = null
watch([openValue, asSheet, () => props.anchor, popup], ([open, sheet, anchor]) => {
	stopAutoUpdate?.()
	stopAutoUpdate = null
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
</script>

<style scoped lang="scss">
// :where() keeps the popover reset at zero specificity so consumers can still style :deep(.popup).
:where(.popup) {
	position: fixed;
	inset: unset;
	margin: 0;
	border: 0;
	padding: 0;
	background: transparent;
	color: inherit;
	overflow-y: auto;
	opacity: 0;
	// Fade in only: the closing fade would need display/overlay allow-discrete, which paints badly
	// in Chromium (see Modal.vue).
	transition: opacity $transition;

	&:popover-open {
		opacity: 1;

		@starting-style {
			opacity: 0;
		}
	}
}
</style>
