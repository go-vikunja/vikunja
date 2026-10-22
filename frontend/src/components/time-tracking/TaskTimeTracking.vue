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

import {useTimeEntries} from '@/composables/useTimeTracking'
import {useDeleteTimeEntryMutation} from '@/client/queries/timeEntries'

import type {TimeEntryResponse as ITimeEntry} from '@/client/queries/timeEntries'

const props = defineProps<{
	taskId: number
}>()

const currentPage = ref(1)
const {entries, totalPages} = useTimeEntries(() => `task_id = ${props.taskId}`, {page: currentPage})
const deleteMutation = useDeleteTimeEntryMutation()
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
