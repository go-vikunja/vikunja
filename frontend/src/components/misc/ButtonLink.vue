<template>
	<BaseButton class="button-link">
		<slot />
	</BaseButton>
</template>

<script setup lang="ts">
import BaseButton from '@/components/base/BaseButton.vue'
</script>

<style lang="scss">
.button-link {
	color: var(--link);

	&:hover {
		color: var(--link-hover);
	}
}
</style>
