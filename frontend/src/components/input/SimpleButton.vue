<template>
	<BaseButton class="simple-button">
		<slot />
	</BaseButton>
</template>

<script lang="ts" setup>
import BaseButton from '@/components/base/BaseButton.vue'
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
