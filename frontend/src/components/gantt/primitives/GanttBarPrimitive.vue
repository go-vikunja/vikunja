<template>
	<component
		:is="as"
		role="slider"
		tabindex="0"
		:aria-valuemin="ariaMin"
		:aria-valuemax="ariaMax"
		:aria-valuenow="ariaNow"
		:aria-valuetext="ariaValueText"
		:aria-label="ariaLabel"
		:data-state="dataState"
		v-bind="attrs"
		@dblclick="() => props.onDoubleClick?.(props.model)"
		@focus="onFocus"
		@blur="onBlur"
		@keydown="onKeyDown"
	>
		<slot
			:dragging="dragging"
			:selected="selected"
			:focused="focused"
		/>
	</component>
</template>

<script setup lang="ts">
import {computed, useAttrs} from 'vue'
import {useI18n} from 'vue-i18n'
import {useGanttBar, type GanttBarModel} from '@/composables/useGanttBar'

const props = withDefaults(
	defineProps<{
		model: GanttBarModel
		timelineStart: Date
		timelineEnd: Date
		onDoubleClick?: (model: GanttBarModel) => void
		onUpdate?: (id: string, newStart: Date, newEnd: Date) => void
		as?: string
	}>(),
	{
		as: 'g',
		onDoubleClick: undefined,
		onUpdate: undefined,
	},
)
const attrs = useAttrs()
const {t} = useI18n({useScope: 'global'})

const {
	dragging,
	selected,
	focused,
	onFocus,
	onBlur,
	onKeyDown,
	// eslint-disable-next-line vue/no-setup-props-reactivity-loss
} = useGanttBar({
	model: props.model,
	timelineStart: props.timelineStart,
	timelineEnd: props.timelineEnd,
	onUpdate: props.onUpdate,
})
const ariaMin = computed(() => props.timelineStart.valueOf())
const ariaMax = computed(() => props.timelineEnd.valueOf())
const ariaNow = computed(() => props.model.start.valueOf())
const ariaValueText = computed(() => `${props.model.start.toLocaleString()} – ${props.model.end.toLocaleString()}`)
const ariaLabel = computed(() =>
	props.model.meta?.label
		? t('project.gantt.taskAriaLabel', { task: props.model.meta.label })
		: t('project.gantt.taskAriaLabelById', { id: props.model.id }),
)
const dataState = computed(() =>
	[
		dragging.value && 'dragging',
		selected.value && 'selected',
		focused.value && 'focused',
	]
		.filter(Boolean)
		.join(' '),
)
</script>
