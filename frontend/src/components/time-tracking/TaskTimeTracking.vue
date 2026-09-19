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
			hide-label-column
			@edit="editingEntry = $event"
			@delete="onDelete"
		/>
	</div>
</template>

<script setup lang="ts">
import {ref, computed} from 'vue'

import TimeEntryForm from '@/components/time-tracking/TimeEntryForm.vue'
import TimeEntryList from '@/components/time-tracking/TimeEntryList.vue'

import {useTimeEntries} from '@/composables/useTimeTracking'
import {useDeleteTimeEntryMutation} from '@/client/queries/timeEntries'

import type {TimeEntryResponse as ITimeEntry} from '@/client/queries/timeEntries'

const props = defineProps<{
	taskId: number
}>()

const {entries} = useTimeEntries(() => `task_id = ${props.taskId}`)
const deleteMutation = useDeleteTimeEntryMutation()
const editingEntry = ref<ITimeEntry | null>(null)
const showForm = ref(false)

// Like related tasks: the form is implicit when empty, otherwise behind the +.
const formVisible = computed(() => entries.value.length === 0 || showForm.value || editingEntry.value !== null)

async function onSaved() {
	editingEntry.value = null
	showForm.value = false
}

function onDelete(id: number) {
	deleteMutation.mutate(id)
}
</script>
