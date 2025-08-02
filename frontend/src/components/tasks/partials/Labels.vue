<template>
	<div class="label-wrapper">
		<XLabel
			v-for="label in displayLabels"
			:key="label.id"
			:label="label"
		/>
	</div>
</template>

<script setup lang="ts">
import {computed} from 'vue'

import type {ILabel} from '@/modelTypes/ILabel'
import XLabel from '@/components/tasks/partials/Label.vue'

const props = defineProps<{
	labels: ILabel[],
}>()

const displayLabels = computed(() =>
	Array.from(new Map(props.labels.map(label => [label.id, label])).values())
		.sort((a, b) => a.title.localeCompare(b.title)),
)
</script>

<style lang="scss" scoped>
.label-wrapper {
	display: inline;
	
	:deep(.tag) {
		margin-block-start: .125rem;
		margin-block-end: .125rem;
	}
}
</style>
