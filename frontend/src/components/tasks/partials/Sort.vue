<template>
	<BaseButton :aria-label="label ? $t('project.table.sortBy', {column: label}) : $t('project.list.sort')">
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
	label?: string
}>(), {
	order: 'none',
	label: undefined,
})
</script>

<style lang="scss" scoped>
.sort__icon {
	color: var(--grey-400);
}
</style>
