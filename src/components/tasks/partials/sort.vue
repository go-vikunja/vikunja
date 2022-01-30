<template>
	<BaseButton>
		<icon icon="sort-up" v-if="order === 'asc'"/>
		<icon icon="sort-up" v-else-if="order === 'desc'" rotation="180"/>
		<icon icon="sort" v-else/>
	</BaseButton>
</template>

<script setup lang="ts">
import {PropType} from 'vue'
import BaseButton from '@/components/base/BaseButton.vue'

type Order = 'asc' | 'desc' | 'none'

defineProps({
	order: {
		type: String as PropType<Order>,
		default: 'none',
	},
})
</script>
