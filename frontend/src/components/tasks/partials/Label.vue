<script setup lang="ts">
import type {ILabel} from '@/modelTypes/ILabel'
import {useLabelStyles} from '@/composables/useLabelStyles'

const props = withDefaults(defineProps<{
	label: ILabel
	clickable?: boolean
}>(), {
	clickable: false,
})

const {getLabelStyles} = useLabelStyles()
</script>

<template>
	<RouterLink
		v-if="clickable"
		:key="label.id"
		:to="{name: 'home', query: {labels: label.id.toString()}}"
		:style="getLabelStyles(label)"
		class="tag tag-clickable"
		@click.stop
	>
		<span>{{ label.title }}</span>
	</RouterLink>
	<span
		v-else
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
	
	&.tag-clickable {
		cursor: pointer;
		transition: opacity $transition;
		
		&:hover {
			opacity: 0.8;
		}
	}
}
</style>
