<template>
	<div class="label-wrapper">
		<span
			:key="label.id"
			:style="{'background': label.hexColor, 'color': label.textColor}"
			class="tag"
			v-for="label in labels">
			<span>{{ label.title }}</span>
		</span>
	</div>
</template>

<script setup lang="ts">
import type { PropType } from 'vue'
import type { ILabel } from '@/models/label'

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

.tag {
	& + & {
		margin-left: 0.5rem;
	}
}
</style>