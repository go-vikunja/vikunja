<template>
	<div
		ref="dropdown"
		class="dropdown"
		role="menu"
		@pointerenter="initialMount = true"
	>
		<slot
			name="trigger"
			:close="close"
			:toggle-open="toggleOpen"
			:open="open"
		>
			<BaseButton
				class="dropdown-trigger is-flex"
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
import {ref, nextTick, watch, computed} from 'vue'
import {onClickOutside} from '@vueuse/core'
import {computePosition, autoPlacement, offset, shift} from '@floating-ui/dom'
import type {IconProp} from '@fortawesome/fontawesome-svg-core'

import CustomTransition from '@/components/misc/CustomTransition.vue'
import BaseButton from '@/components/base/BaseButton.vue'

withDefaults(defineProps<{
	triggerIcon?: IconProp
}>(), {
	triggerIcon: 'ellipsis-h',
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
	'default': () => void
}>()

const initialMount = ref(false)
const open = ref(false)
const dropdown = ref<HTMLElement>()
const dropdownMenu = ref<HTMLElement>()
const dropdownPosition = ref({x: 0, y: 0})

function close() {
	open.value = false
}

async function updatePosition() {
	if (!dropdown.value || !dropdownMenu.value) {
		return
	}

	await nextTick()

	const {x, y} = await computePosition(dropdown.value, dropdownMenu.value, {
		placement: 'bottom-end',
		strategy: 'absolute',
		middleware: [
			offset(4),
			autoPlacement({
				allowedPlacements: ['bottom-end', 'top-end', 'bottom-start', 'top-start'],
				padding: 8,
			}),
			shift({padding: 8}),
		],
	})

	dropdownPosition.value = {x, y}
}

const dropdownMenuStyle = computed(() => ({
	left: `${dropdownPosition.value.x}px`,
	top: `${dropdownPosition.value.y}px`,
}))

function toggleOpen() {
	open.value = !open.value
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

.dropdown-menu {
	min-inline-size: 12rem;
	position: fixed;
	z-index: 20;
	display: block;
}

.dropdown-content {
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
