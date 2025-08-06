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
				class="dropdown-menu"
			>
				<div class="dropdown-content">
					<slot :close="close" />
				</div>
			</div>
		</CustomTransition>
	</div>
</template>

<script setup lang="ts">
import {ref} from 'vue'
import {onClickOutside} from '@vueuse/core'
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

function close() {
	open.value = false
}

function toggleOpen() {
	open.value = !open.value
}

const dropdown = ref()
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
	padding-block-start: 4px;
	position: absolute;
	inset-block-start: 100%;
	z-index: 20;
	display: block;
	inset-inline-start: auto;
	inset-inline-end: 0;
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
