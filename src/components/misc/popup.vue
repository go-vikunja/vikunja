<template>
	<slot name="trigger" :isOpen="open" :toggle="toggle"></slot>
	<div class="popup" :class="{'is-open': open, 'has-overflow': props.hasOverflow && open}" ref="popup">
		<slot name="content" :isOpen="open"/>
	</div>
</template>

<script setup>
import {closeWhenClickedOutside} from '@/helpers/closeWhenClickedOutside'
import {onBeforeUnmount, onMounted, ref} from 'vue'

const open = ref(false)
const popup = ref(null)

const toggle = () => {
	open.value = !open.value
}

const props = defineProps({
	hasOverflow: {
		type: Boolean,
		default: false,
	},
})

function hidePopup(e) {
	if (!open.value) {
		return
	}

	// we actually want to use popup.$el, not its value.
	// eslint-disable-next-line vue/no-ref-as-operand
	closeWhenClickedOutside(e, popup.value, () => {
		open.value = false
	})
}

onMounted(() => {
	document.addEventListener('click', hidePopup)
})

onBeforeUnmount(() => {
	document.removeEventListener('click', hidePopup)
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

	&.is-open {
		opacity: 1;
		height: auto;
	}
}
</style>
