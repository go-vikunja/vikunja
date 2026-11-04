<template>
	<div
		ref="dropdown"
		class="dropdown"
		@pointerenter="initialMount = true"
		@keydown="onKeydown"
	>
		<slot
			name="trigger"
			:close="close"
			:toggle-open="toggleOpen"
			:open="open"
		>
			<BaseButton
				class="dropdown-trigger is-flex"
				:aria-label="triggerLabel"
				:aria-expanded="open"
				@click="toggleOpen"
			>
				<Icon
					:icon="triggerIcon"
					class="icon"
				/>
			</BaseButton>
		</slot>

		<CustomTransition name="fade">
			<div
				v-if="initialMount || open"
				v-show="open"
				ref="dropdownMenu"
				class="dropdown-menu"
				:style="dropdownMenuStyle"
			>
				<div class="dropdown-content">
					<slot :close="close" />
				</div>
			</div>
		</CustomTransition>
	</div>
</template>

<script setup lang="ts">
import {ref, nextTick, watch, computed, type PropType} from 'vue'
import {onClickOutside} from '@vueuse/core'
import {computePosition, autoPlacement, offset, shift, size} from '@floating-ui/dom'
import type {IconProp} from '@fortawesome/fontawesome-svg-core'

import CustomTransition from '@/components/misc/CustomTransition.vue'
import BaseButton from '@/components/base/BaseButton.vue'

defineProps({
	triggerIcon: {type: [String, Array, Object] as PropType<IconProp>, default: 'ellipsis-h'},
	triggerLabel: {type: String, default: undefined},
})

const emit = defineEmits<{
	'close': [event: PointerEvent]
}>()

defineSlots<{
	'trigger': (props: {
		close: () => void,
		toggleOpen: () => void, 
		open: boolean
	}) => void,
	'default': (props: {close: () => void}) => void
}>()

const initialMount = ref(false)
const open = ref(false)

const dropdown = ref<HTMLElement>()
const dropdownMenu = ref<HTMLElement>()
const dropdownPosition = ref({x: 0, y: 0})
const dropdownMenuOffset = computed(() => 4)

function close() {
	open.value = false
}

// Menus in a dialog render in the top layer, above the navbar.
function fixedNavbarBottom(trigger: HTMLElement): number {
	const navbar = document.querySelector<HTMLElement>('.navbar')
	if (!navbar || trigger.closest('dialog[open]') || getComputedStyle(navbar).position !== 'fixed') {
		return 0
	}
	return navbar.getBoundingClientRect().bottom
}

async function updatePosition() {
	if (!dropdown.value || !dropdownMenu.value) {
		return
	}

	await nextTick()

	const padding = {top: fixedNavbarBottom(dropdown.value) + 8, right: 8, bottom: 8, left: 8}

	const {x, y} = await computePosition(dropdown.value, dropdownMenu.value, {
		placement: 'bottom-end',
		strategy: 'absolute',
		middleware: [
			offset(dropdownMenuOffset.value),
			autoPlacement({
				allowedPlacements: ['bottom-end', 'top-end', 'bottom-start', 'top-start'],
				padding,
			}),
			shift({padding}),
			size({
				padding,
				// Set synchronously so floating-ui re-measures the capped menu before placing it.
				apply({availableHeight, elements}) {
					elements.floating.style.setProperty('--dropdown-max-height', `${availableHeight}px`)
				},
			}),
		],
	})

	dropdownPosition.value = {x, y}
}

const dropdownMenuStyle = computed(() => ({
	left: `${dropdownPosition.value.x}px`,
	top: `${dropdownPosition.value.y}px`,
	'--hover-offset': `${dropdownMenuOffset.value}px`,
}))

function toggleOpen() {
	open.value = !open.value
}

function onKeydown(e: KeyboardEvent) {
	if (e.key !== 'Escape' || !open.value) {
		return
	}
	e.stopPropagation()
	close()
	focusTrigger()
}

// Return focus to the trigger, which is the first focusable element that lives
// outside the popup menu.
function focusTrigger() {
	const focusables = dropdown.value?.querySelectorAll<HTMLElement>('button, a[href], input, [tabindex]')
	for (const el of focusables ?? []) {
		if (!el.closest('.dropdown-menu')) {
			el.focus()
			return
		}
	}
}

watch(open, (isOpen) => {
	if (isOpen) {
		updatePosition()
	}
})

onClickOutside(dropdown, (e) => {
	if (!open.value) {
		return
	}
	close()
	emit('close', e)
})
</script>

<style lang="scss" scoped>
.dropdown {
	display: inline-flex;
	position: relative;
}

.dropdown-menu::before {
  content: "";
  position: absolute;
  inset: calc(var(--hover-offset) * -1);
}

.dropdown-menu {
	min-inline-size: 12rem;
	position: absolute;
	z-index: 20;
	display: block;
}

.dropdown-content {
	max-block-size: var(--dropdown-max-height, none);
	overflow-y: auto;
	background-color: var(--scheme-main);
	border-radius: $radius;
	padding-block-end: .5rem;
	padding-block-start: .5rem;
	box-shadow: var(--shadow-md);
}

.dropdown-divider {
	background-color: var(--border-light);
	border: none;
	display: block;
	block-size: 1px;
	margin: 0.5rem 0;
}
</style>
