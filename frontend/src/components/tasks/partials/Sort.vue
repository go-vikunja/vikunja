<template>
	<BaseButton>
		<Icon
			v-if="order === 'asc'"
			icon="sort-up"
			class="sort__icon"
		/>
		<Icon
			v-else-if="order === 'desc'"
			icon="sort-up"
			rotation="180"
			class="sort__icon"
		/>
		<Icon
			v-else
			icon="sort"
			class="sort__icon"
		/>
	</BaseButton>
</template>

<script setup lang="ts">
import BaseButton from '@/components/base/BaseButton.vue'

withDefaults(defineProps<{
	order?: 'asc' | 'desc' | 'none'
}>(), {
	order: 'none',
})
</script>

<style lang="scss" scoped>
.sort__icon {
	color: var(--grey-400);
}
</style>
