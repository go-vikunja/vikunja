<template>
	<slot
		name="trigger"
		:is-open="open"
		:toggle="toggle"
		:close="close"
	/>
	<div
		ref="popup"
		class="popup"
		:class="{
			'is-open': open,
			'has-overflow': props.hasOverflow && open
		}"
	>
		<slot
			name="content"
			:is-open="open"
			:toggle="toggle"
			:close="close"
		/>
	</div>
</template>

<script setup lang="ts">
import {ref, watch} from 'vue'
import {onClickOutside} from '@vueuse/core'

const props = defineProps({
	hasOverflow: {
		type: Boolean,
		default: false,
	},
	open: {
		type: Boolean,
		default: false,
	},
})

watch(
	() => props.open,
	nowOpen => {
		open.value = nowOpen
	},
)

const emit = defineEmits(['close'])

const open = ref(false)
const popup = ref<HTMLElement | null>(null)

function close() {
	open.value = false
	emit('close')
}

function toggle() {
	open.value = !open.value
}

onClickOutside(popup, () => {
	if (!open.value) {
		return
	}
	close()
})
</script>

<style scoped lang="scss">
.popup {
	transition: opacity $transition;
	opacity: 0;
	height: 0;
	overflow: hidden;
	position: absolute;
	top: 1rem;
	z-index: 100;

	&.is-open {
		opacity: 1;
		height: auto;
	}
}
</style>
