<template>
	<BaseButton>
		<PhSortAscending
			v-if="order === 'asc'"
			class="sort__icon"
		/>
		<PhSortDescending
			v-else-if="order === 'desc'"
			class="sort__icon"
		/>
		<PhArrowsDownUp
			v-else
			class="sort__icon"
		/>
	</BaseButton>
</template>

<script setup lang="ts">
import {PhSortAscending, PhSortDescending, PhArrowsDownUp} from '@phosphor-icons/vue'
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
