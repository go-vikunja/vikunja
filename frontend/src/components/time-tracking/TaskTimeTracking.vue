<template>
	<div class="task-time-tracking">
		<XButton
			v-if="entries.length > 0"
			v-tooltip="$t('timeTracking.logTime')"
			v-cy="'addTaskTimeEntry'"
			:aria-label="$t('timeTracking.logTime')"
			class="is-pulled-right d-print-none"
			:class="{'is-active': showForm}"
			variant="secondary"
			icon="plus"
			:shadow="false"
			@click="showForm = !showForm"
		/>
		<h3 class="title is-5">
			{{ $t('timeTracking.title') }}
		</h3>
		<div
			v-if="runningTimer"
			v-cy="'taskRunningTimer'"
			class="running-timer"
		>
			<span class="has-text-grey">{{ $t('timeTracking.running') }}</span>
			<span class="running-timer__elapsed">{{ elapsed }}</span>
			<span
				v-if="runningTimer.comment"
				class="running-timer__comment has-text-grey"
			>{{ runningTimer.comment }}</span>
			<XButton
				v-cy="'stopTaskTimer'"
				class="running-timer__stop"
				variant="secondary"
				icon="stop"
				:shadow="false"
				:loading="stopMutation.isPending.value"
				@click="stop"
			>
				{{ $t('timeTracking.stop') }}
			</XButton>
		</div>
		<TimeEntryForm
			v-if="formVisible"
			:task-id="taskId"
			:entry="editingEntry"
			:recent-entries="entries"
			@saved="onSaved"
			@cancel="editingEntry = null"
		/>
		<TimeEntryList
			class="mbs-4"
			:entries="entries"
			:card="false"
			:empty-text="$t('timeTracking.list.emptyTask')"
			:paged="totalPages > 1"
			hide-label-column
			@edit="editingEntry = $event"
			@delete="onDelete"
		/>
		<PaginationEmit
			v-if="totalPages > 1"
			:total-pages="totalPages"
			:current-page="currentPage"
			@pageChanged="currentPage = $event"
		/>
	</div>
</template>

<script setup lang="ts">
import {ref, computed, watch} from 'vue'

import PaginationEmit from '@/components/misc/PaginationEmit.vue'
import TimeEntryForm from '@/components/time-tracking/TimeEntryForm.vue'
import TimeEntryList from '@/components/time-tracking/TimeEntryList.vue'

import {useElapsed, useTimeEntries, useTimeTracking} from '@/composables/useTimeTracking'
import {useDeleteTimeEntryMutation, useStopTimerMutation} from '@/client/queries/timeEntries'

import type {TimeEntryResponse as ITimeEntry} from '@/client/queries/timeEntries'

const props = defineProps<{
	taskId: number
}>()

const currentPage = ref(1)
const {entries, totalPages} = useTimeEntries(() => `task_id = ${props.taskId}`, {page: currentPage})
const deleteMutation = useDeleteTimeEntryMutation()
const stopMutation = useStopTimerMutation()
const {activeTimer} = useTimeTracking()
const runningTimer = computed(() => activeTimer.value?.task_id === props.taskId ? activeTimer.value : null)
const elapsed = useElapsed(() => runningTimer.value?.start_time)
const editingEntry = ref<ITimeEntry | null>(null)
const showForm = ref(false)

watch(() => props.taskId, () => {
	currentPage.value = 1
})

// Like related tasks: the form is implicit when empty, otherwise behind the +.
const formVisible = computed(() => entries.value.length === 0 || showForm.value || editingEntry.value !== null)

function onSaved() {
	editingEntry.value = null
	showForm.value = false
}

function stop() {
	if (!stopMutation.isPending.value) {
		stopMutation.mutate(undefined)
	}
}

async function onDelete(entry: ITimeEntry) {
	if (deleteMutation.isPending.value) {
		return
	}
	try {
		await deleteMutation.mutateAsync({
			id: entry.id,
			taskId: entry.task_id,
		})
	} catch {
		return
	}
	currentPage.value = Math.min(currentPage.value, Math.max(1, totalPages.value))
}
</script>

<style lang="scss" scoped>
.running-timer {
	display: flex;
	align-items: center;
	gap: .5rem;
	margin-block-end: 1rem;
}

.running-timer__elapsed {
	color: var(--primary);
	font-variant-numeric: tabular-nums;
	font-weight: 600;
}

.running-timer__comment {
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
}

.running-timer__stop {
	margin-inline-start: auto;
}
</style>
