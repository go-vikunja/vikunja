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
import type {ILabel} from '@/modelTypes/ILabel'

import XLabel from '@/components/tasks/partials/Label.vue'

defineProps<{
	labels: ILabel[],
}>()
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