<template>
	<div class="dropdown" ref="dropdown">
		<slot name="trigger" :close="close" :toggleOpen="toggleOpen" :open="open">
			<BaseButton class="dropdown-trigger is-flex" @click="toggleOpen">
				<icon :icon="triggerIcon" class="icon"/>
			</BaseButton>
		</slot>

		<CustomTransition name="fade">
			<div class="dropdown-menu" v-if="open">
				<div class="dropdown-content">
					<slot :close="close"></slot>
				</div>
			</div>
		</CustomTransition>
	</div>
</template>

<script setup lang="ts">
import {ref, type PropType} from 'vue'
import {onClickOutside} from '@vueuse/core'
import type {IconProp} from '@fortawesome/fontawesome-svg-core'

import CustomTransition from '@/components/misc/CustomTransition.vue'
import BaseButton from '@/components/base/BaseButton.vue'

defineProps({
	triggerIcon: {
		type: String as PropType<IconProp>,
		default: 'ellipsis-h',
	},
})
const emit = defineEmits(['close'])

const open = ref(false)

function close() {
	open.value = false
}

function toggleOpen() {
	open.value = !open.value
}

const dropdown = ref()
onClickOutside(dropdown, (e: Event) => {
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
	min-width: 12rem;
	padding-top: 4px;
	position: absolute;
	top: 100%;
	z-index: 20;
	display: block;
	left: auto;
	right: 0;
}

.dropdown-content {
	background-color: var(--scheme-main);
	border-radius: $radius;
	padding-bottom: .5rem;
	padding-top: .5rem;
	box-shadow: var(--shadow-md);
}

.dropdown-divider {
	background-color: var(--border-light);
	border: none;
	display: block;
	height: 1px;
	margin: 0.5rem 0;
}
</style>
