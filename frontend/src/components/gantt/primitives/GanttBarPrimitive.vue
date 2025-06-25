<template>
	<div role="gridcell">
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
			@pointerdown="onPointerDown"
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
	</div>
</template>

<script setup lang="ts">
import {computed,useAttrs} from 'vue'
import {useGanttBar, type GanttBarModel} from '@/composables/useGanttBar'
const props = withDefaults(defineProps<{
  model: GanttBarModel
  timelineStart: Date
  timelineEnd: Date
  onMove: (id: string, newStart: Date, newEnd: Date) => void
  onDoubleClick?: (model: GanttBarModel) => void
  as?: string
}>(), { as: 'div', onDoubleClick: undefined })
const attrs = useAttrs()
// eslint-disable-next-line vue/no-setup-props-reactivity-loss
const {dragging, selected, focused, onPointerDown, onFocus, onBlur, onKeyDown} = useGanttBar({
	model: props.model,
	timelineStart: props.timelineStart,
	timelineEnd: props.timelineEnd,
	onMove: props.onMove,
})
const ariaMin = computed(()=>props.timelineStart.valueOf())
const ariaMax = computed(()=>props.timelineEnd.valueOf())
const ariaNow = computed(()=>props.model.start.valueOf())
const ariaValueText = computed(()=>`${props.model.start.toLocaleString()} – ${props.model.end.toLocaleString()}`)
const ariaLabel = computed(()=>`Task ${props.model.id}`)
const dataState = computed(()=>[dragging.value&&'dragging', selected.value&&'selected', focused.value&&'focused'].filter(Boolean).join(' '))
</script>
