<template>
	<div class="dropdown is-right is-active" ref="dropdown">
		<slot name="trigger" :close="close" :toggleOpen="toggleOpen">
			<BaseButton class="dropdown-trigger is-flex" @click="toggleOpen">
				<icon :icon="triggerIcon" class="icon"/>
			</BaseButton>
		</slot>

		<transition name="fade">
			<div class="dropdown-menu" v-if="open">
				<div class="dropdown-content">
					<slot :close="close"></slot>
				</div>
			</div>
		</transition>
	</div>
</template>

<script setup lang="ts">
import {ref} from 'vue'
import {onClickOutside} from '@vueuse/core'

import BaseButton from '@/components/base/BaseButton.vue'

defineProps({
	triggerIcon: {
		type: String,
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
.dropdown-menu  .dropdown-content {
	box-shadow: var(--shadow-md);
}
</style>