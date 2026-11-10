<template>
	<component
		:is="is"
		class="switch-view"
	>
		<slot />
	</component>
</template>

<script setup lang="ts">
import type {Component} from 'vue'

withDefaults(defineProps<{
	is?: string | Component,
}>(), {
	is: 'div',
})
</script>

<style lang="scss" scoped>
.switch-view {
	background: var(--white);
	display: inline-flex;
	border-radius: $radius;
	font-size: .75rem;
	box-shadow: var(--shadow-sm);
	padding: .5rem;
}

::v-slotted(.switch-view-button) {
	padding: .25rem .5rem;
	display: block;
	white-space: nowrap;
	border-radius: $radius;
	transition: all 100ms;

	&:not(:last-child) {
		margin-inline-end: .5rem;
	}

	&:hover {
		color: var(--switch-view-color);
		background: var(--switch-view-active-background);
	}

	&.is-active {
		color: var(--switch-view-color);
		background: var(--switch-view-active-background);
		font-weight: bold;
		box-shadow: var(--shadow-xs);
	}
}
</style>
