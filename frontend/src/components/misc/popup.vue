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
			'has-overflow': props.hasOverflow && openValue
		}"
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

const emit = defineEmits(['close'])

watch(
	() => props.open,
	nowOpen => {
		openValue.value = nowOpen
	},
)

const openValue = ref(false)
const popup = ref<HTMLElement | null>(null)

function close() {
	openValue.value = false
	emit('close')
}

function toggle() {
	openValue.value = !openValue.value
}

onClickOutside(popup, () => {
	if (!openValue.value) {
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
