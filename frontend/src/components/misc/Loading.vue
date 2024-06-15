<template>
	<div
		class="loader-container is-loading"
		:class="{'is-small': variant === 'small'}"
	/>
</template>

<script lang="ts" setup>
withDefaults(defineProps<{
	variant?: 'default' | 'small'
}>(), {
	variant: 'default',
})
</script>

<style scoped lang="scss">
.loader-container {
	height: 100%;
	min-height: 200px;
	width: 100%;
	min-width: 600px;
	max-width: 100vw;

	&.is-loading-small {
		min-height: 50px;
		min-width: 100px;
	}

	&.is-small {
		min-width: 100%;
		height: 150px;

		&.is-loading::after {
			width: 3rem;
			height: 3rem;
			top: calc(50% - 1.5rem);
			left: calc(50% - 1.5rem);
			border-width: 3px;
		}
	}
}
</style>