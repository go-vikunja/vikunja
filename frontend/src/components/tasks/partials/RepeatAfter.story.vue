<script setup lang="ts">
import {provide, ref} from 'vue'
import {QueryClient, VUE_QUERY_CLIENT} from '@tanstack/vue-query'

import RepeatAfter, {type RepeatSaveResult} from './RepeatAfter.vue'
import type {PreviewAnswer, Task, TaskRepeat} from '@/client/generated'
import {client} from '@/client/generated/client.gen'
import {isRefusal, previewFixtures} from '@/client/queries/occurrencePreview.fixtures'
import type {TaskInput} from '@/client/queries/taskMutations'

function answerOf(row: string): PreviewAnswer {
	const answer = previewFixtures.find(fixture => fixture.row === row)?.answer
	if (!answer || isRefusal(answer)) throw new Error(`no answer for ${row}`)
	return answer
}

// There is no server behind the stories: every preview gets a recorded answer, dated or undated.
client.setConfig({
	fetch: async (request: RequestInfo | URL) => {
		const body = request instanceof Request ? await request.clone().json() : {}
		const answer = answerOf(body.due_date ? 'P-10' : 'P-1')
		return new Response(JSON.stringify(answer), {headers: {'Content-Type': 'application/json'}})
	},
})
provide(VUE_QUERY_CLIENT, new QueryClient())

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

// Stands in for the task view: the save lands, without a server choosing a due date.
async function save(key: string, value: TaskInput): Promise<RepeatSaveResult> {
	models.value[key] = {...value, repeat: value.repeat ?? undefined} as Task
	return {ok: true, outcome: {}}
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
						:save="value => save(`${state.title}-${width}`, value)"
					/>
				</div>
			</Variant>
		</template>
	</Story>
</template>
