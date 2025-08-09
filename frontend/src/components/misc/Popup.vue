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
	ignoreClickClasses?: string[]
}>(), {
	hasOverflow: false,
	open: false,
	ignoreClickClasses: () => [],
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

function toggle() {
	openValue.value = !openValue.value
	emit('update:open', openValue.value)
	return openValue.value
}

const popup = ref<HTMLElement | null>(null)

onClickOutside(popup, (event) => {
	const target = event.target as HTMLElement
	// Check if the click target has any of the ignored classes
	if (target?.classList && props.ignoreClickClasses.some(className => target.classList.contains(className))) {
		return
	}
	close()
})
</script>

<style scoped lang="scss">
.popup {
	transition: opacity $transition;
	opacity: 0;
	block-size: 0;
	overflow: hidden;
	position: absolute;
	inset-block-start: 1rem;
	z-index: 100;

	&.is-open {
		opacity: 1;
		block-size: auto;
	}
}
</style>
