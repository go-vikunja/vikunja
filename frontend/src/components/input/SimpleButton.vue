<template>
	<BaseButton
		ref="baseButton"
		class="simple-button"
	>
		<slot />
	</BaseButton>
</template>

<script lang="ts" setup>
import {ref} from 'vue'
import BaseButton from '@/components/base/BaseButton.vue'

const baseButton = ref<InstanceType<typeof BaseButton> | null>(null)

function focus() {
	baseButton.value?.focus()
}

defineExpose({
	focus,
})
</script>

<style lang="scss" scoped>
.simple-button {
	color: var(--text);
	padding: .25rem .5rem;
	transition: background-color $transition;
	border-radius: $radius;
	display: block;
	margin: .1rem 0;
	inline-size: 100%;
	text-align: start;

	&:hover {
		background: var(--white);
	}
}
</style>
