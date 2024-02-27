<template>
	<div class="label-wrapper">
		<XLabel
			v-for="label in labels"
			:key="label.id"
			:label="label"
		/>
	</div>
</template>

<script setup lang="ts">
import type {PropType} from 'vue'
import type {ILabel} from '@/modelTypes/ILabel'

import XLabel from '@/components/tasks/partials/label.vue'

defineProps({
	labels: {
		type: Array as PropType<ILabel[]>,
		required: true,
	},
})
</script>

<style lang="scss" scoped>
.label-wrapper {
	display: inline;
	
	:deep(.tag) {
		margin-top: .125rem;
		margin-bottom: .125rem;
	}
}
</style>