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
	block-size: 100%;
	min-block-size: 200px;
	inline-size: 100%;
	min-inline-size: 600px;
	max-inline-size: 100vw;

	&.is-loading-small {
		min-block-size: 50px;
		min-inline-size: 100px;
	}

	&.is-small {
		min-inline-size: 100%;
		block-size: 150px;

		&.is-loading::after {
			inline-size: 3rem;
			block-size: 3rem;
			inset-block-start: calc(50% - 1.5rem);
			inset-inline-start: calc(50% - 1.5rem);
			border-width: 3px;
		}
	}
}
</style>
