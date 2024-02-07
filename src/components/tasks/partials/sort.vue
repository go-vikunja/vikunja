<template>
	<BaseButton>
		<icon
			v-if="order === 'asc'"
			icon="sort-up"
		/>
		<icon
			v-else-if="order === 'desc'"
			icon="sort-up"
			rotation="180"
		/>
		<icon
			v-else
			icon="sort"
		/>
	</BaseButton>
</template>

<script setup lang="ts">
import type {PropType} from 'vue'
import BaseButton from '@/components/base/BaseButton.vue'

type Order = 'asc' | 'desc' | 'none'

defineProps({
	order: {
		type: String as PropType<Order>,
		default: 'none',
	},
})
</script>
