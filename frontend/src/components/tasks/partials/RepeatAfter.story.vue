<script setup lang="ts">
import {ref} from 'vue'

import RepeatAfter from './RepeatAfter.vue'
import type {Task, TaskRepeat} from '@/client/generated'
import type {TaskInput} from '@/client/queries/taskMutations'

const dueDate = new Date()
dueDate.setDate(dueDate.getDate() + 3)
dueDate.setHours(12, 0, 0, 0)

function task(repeat: TaskRepeat | null, due: Date | null = dueDate): Task {
	return {id: 1, title: 'Water the plants', due_date: due?.toISOString(), repeat: repeat ?? undefined, repeats_from_current_date: false}
}

const widths = [290, 375]

const states = [
	{title: 'Does not repeat', value: () => task(null)},
	{title: 'Weekly on Mon, Wed, Fri', value: () => task({freq: 'weekly', by_day: ['mo', 'we', 'fr']})},
	{title: 'Not editable here', value: () => task({freq: 'yearly', by_month: [3, 9], by_day: ['1mo']})},
	{title: 'Set elsewhere, read-only', value: () => task({freq: 'yearly', interval: 1, read_only: true})},
	{title: 'No due date', value: () => task(null, null)},
]

const models = ref<Record<string, Task>>(Object.fromEntries(states.flatMap(state => widths.map(width => [`${state.title}-${width}`, state.value()]))))

function save(key: string, value: TaskInput) {
	models.value[key] = {...value, repeat: value.repeat ?? undefined} as Task
}
</script>

<template>
	<Story>
		<template
			v-for="state in states"
			:key="state.title"
		>
			<Variant
				v-for="width in widths"
				:key="width"
				:title="`${state.title} (${width} px)`"
			>
				<div :style="{inlineSize: `${width}px`, minBlockSize: '36rem'}">
					<RepeatAfter
						:model-value="models[`${state.title}-${width}`]"
						@update:modelValue="save(`${state.title}-${width}`, $event)"
					/>
				</div>
			</Variant>
		</template>
	</Story>
</template>
