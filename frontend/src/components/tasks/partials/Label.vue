<script setup lang="ts">
import type {ILabel} from '@/modelTypes/ILabel'

defineProps<{
	label: ILabel
}>()
</script>

<template>
	<span
		:key="label.id"
		:style="{'background': label.hexColor || 'var(--grey-200)', 'color': label.textColor}"
		class="tag"
	>
		<span>{{ label.title }}</span>
	</span>
</template>

<style scoped lang="scss">
.tag {
	& + & {
		margin-left: 0.5rem;
	}
}
</style>
