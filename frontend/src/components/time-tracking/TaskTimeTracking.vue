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
import {ref, computed, watch} from 'vue'

import TimeEntryForm from '@/components/time-tracking/TimeEntryForm.vue'
import TimeEntryList from '@/components/time-tracking/TimeEntryList.vue'

import {useTimeEntryService} from '@/services/timeEntry'
import {useTimeTrackingStore} from '@/stores/timeTracking'

import type {ITimeEntry} from '@/modelTypes/ITimeEntry'

const props = defineProps<{
	taskId: number
}>()

const timeTrackingStore = useTimeTrackingStore()
const entries = ref<ITimeEntry[]>([])
const editingEntry = ref<ITimeEntry | null>(null)
const showForm = ref(false)

// Like related tasks: the form is implicit when empty, otherwise behind the +.
const formVisible = computed(() => entries.value.length === 0 || showForm.value || editingEntry.value !== null)

async function load() {
	const {items} = await useTimeEntryService().getAll({
		filter: `task_id = ${props.taskId}`,
		perPage: 250,
	})
	entries.value = items
}

async function onSaved() {
	editingEntry.value = null
	showForm.value = false
	await load()
}

async function onDelete(id: number) {
	await timeTrackingStore.removeEntry(id)
	await load()
}

watch(() => props.taskId, load, {immediate: true})
// The header badge can start/stop the timer without going through this form;
// reload so the row reflects the stop (its new end time).
watch(() => timeTrackingStore.activeTimer, load)
</script>
