<template>
	<div class="label-wrapper">
		<XLabel
			v-for="label in labels"
			:label="label"
			:key="label.id"
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
}
</style>