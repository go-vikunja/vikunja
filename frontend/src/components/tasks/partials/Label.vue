<script setup lang="ts">
import type {ILabel} from '@/modelTypes/ILabel'
import {useLabelStyles} from '@/composables/useLabelStyles'

defineProps<{
	label: ILabel
}>()

const {getLabelStyles} = useLabelStyles()
</script>

<template>
	<span
		:key="label.id"
		:style="getLabelStyles(label)"
		class="tag"
	>
		<span>{{ label.title }}</span>
	</span>
</template>

<style scoped lang="scss">
.tag {
	& + & {
		margin-inline-start: 0.5rem;
	}
}
</style>
