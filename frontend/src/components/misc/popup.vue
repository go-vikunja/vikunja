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
import {ref, watchEffect} from 'vue'
import {onClickOutside} from '@vueuse/core'

const props = withDefaults(defineProps<{
	hasOverflow?: boolean
	open?: boolean
}>(), {
	hasOverflow: false,
	open: false,
})

const emit = defineEmits(['close'])

const openValue = ref(props.open)
watchEffect(() => {
	openValue.value = props.open
})

function close() {
	openValue.value = false
	emit('close')
}

function toggle() {
	openValue.value = !openValue.value
}

const popup = ref<HTMLElement | null>(null)

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
