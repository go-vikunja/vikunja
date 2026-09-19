<template>
	<form
		ref="formEl"
		v-cy="'timeEntryForm'"
		class="time-entry-form"
		@submit.prevent="saveEntry"
	>
		<div
			v-if="taskId === undefined"
			class="field-columns"
		>
			<div class="field">
				<label class="label">{{ $t('task.attributes.project') }}</label>
				<ProjectSearch v-model="selectedProject" />
			</div>

			<div class="field">
				<label class="label">{{ $t('timeTracking.form.task') }}</label>
				<Multiselect
					v-model="selectedTask"
					:placeholder="$t('timeTracking.form.taskSearch')"
					:loading="taskQuery.isFetching.value"
					:search-results="foundTasks"
					label="title"
					@search="findTasks"
				>
					<template #searchResult="{option}">
						{{ option.title }}
					</template>
				</Multiselect>
			</div>
		</div>

		<div class="field">
			<label class="label">{{ $t('task.comment.comment') }}</label>
			<input
				v-model="comment"
				v-cy="'timeEntryComment'"
				class="input"
				type="text"
				:placeholder="$t('timeTracking.form.commentPlaceholder')"
			>
		</div>

		<div class="field is-grouped from-to-row">
			<div class="control is-expanded">
				<label class="label">{{ $t('input.datepickerRange.from') }}</label>
				<Datepicker
					v-model="from"
					:show-shortcuts="false"
				/>
			</div>
			<div class="control is-expanded">
				<label class="label">{{ $t('input.datepickerRange.to') }}</label>
				<Datepicker
					v-model="to"
					:show-shortcuts="false"
					:empty-label="$t('misc.notSet')"
				/>
			</div>
			<div class="control">
				<BaseButton
					v-tooltip="$t('timeTracking.form.smartFill')"
					v-cy="'smartFill'"
					class="smart-fill"
					:aria-label="$t('timeTracking.form.smartFill')"
					@click="smartFill"
				>
					<Icon :icon="['far', 'clock']" />
				</BaseButton>
			</div>
		</div>

		<div class="field form-actions">
			<template v-if="isEditing">
				<XButton
					v-cy="'updateTimeEntry'"
					:aria-disabled="!canSubmit || undefined"
					:loading="isSaving"
					@click="saveEntry"
				>
					{{ $t('timeTracking.form.update') }}
				</XButton>
				<XButton
					variant="secondary"
					@click="cancelEdit"
				>
					{{ $t('misc.cancel') }}
				</XButton>
			</template>
			<template v-else>
				<XButton
					v-cy="'saveTimeEntry'"
					:aria-disabled="!canSubmit || undefined"
					:loading="isSaving"
					@click="saveEntry"
				>
					{{ $t('timeTracking.form.save') }}
				</XButton>
				<XButton
					v-cy="'startTimer'"
					variant="secondary"
					:aria-disabled="!canSubmit || undefined"
					:loading="isSaving"
					@click="startTimer"
				>
					{{ $t('timeTracking.form.startTimer') }}
				</XButton>
			</template>
		</div>
	</form>
</template>

<script setup lang="ts">
import {ref, computed, watch, nextTick} from 'vue'

import BaseButton from '@/components/base/BaseButton.vue'
import Multiselect from '@/components/input/Multiselect.vue'
import Datepicker from '@/components/input/Datepicker.vue'
import ProjectSearch from '@/components/tasks/partials/ProjectSearch.vue'

import {useTasks} from '@/composables/useTasks'
import {ensureTask} from '@/client/queries/tasks'
import {smartFillStart} from '@/helpers/time/smartFillStart'
import {useCreateTimeEntryMutation, useUpdateTimeEntryMutation} from '@/client/queries/timeEntries'
import type {UpdateTimeEntryInput} from '@/client/queries/timeEntries'
import type {TimeEntryWritable} from '@/client/generated'
import {parseDateOrNull} from '@/helpers/parseDateOrNull'
import {useAuthStore} from '@/stores/auth'
import {useProjects} from '@/composables/useProjects'

import type {ProjectResponse} from '@/client/queries/projects'
import type {TaskResponse} from '@/client/queries/tasks'
import type {TimeEntryResponse as ITimeEntry} from '@/client/queries/timeEntries'

const props = withDefaults(defineProps<{
	// When set, the entry is locked to this task and the project/task pickers are hidden.
	taskId?: number
	// When set, the form edits this entry (Update + Cancel) instead of creating.
	entry?: ITimeEntry | null
	// Entries the smart-clock looks at to continue from the last one's end.
	recentEntries?: ITimeEntry[]
}>(), {
	taskId: undefined,
	entry: undefined,
	recentEntries: () => [],
})

const emit = defineEmits<{
	saved: []
	cancel: []
}>()

const createMutation = useCreateTimeEntryMutation()
const updateMutation = useUpdateTimeEntryMutation()
const authStore = useAuthStore()
const projectList = useProjects()

const isEditing = computed(() => props.entry != null)

const formEl = ref<HTMLFormElement | null>(null)
const selectedProject = ref<ProjectResponse | null>(null)
const selectedTask = ref<TaskResponse | null>(null)
const from = ref<Date | null>(new Date())
const to = ref<Date | null>(null)
const comment = ref('')
const isSaving = ref(false)

// Task and project are mutually exclusive (XOR) — selecting one clears the other,
// so applyTarget never picks a stale target the user has since changed.
watch(selectedTask, task => {
	if (task !== null) {
		selectedProject.value = null
	}
})
watch(selectedProject, project => {
	if (project !== null) {
		selectedTask.value = null
	}
})

const taskSearch = ref('')
const taskQuery = useTasks(
	() => ({project: selectedProject.value?.id, params: {q: taskSearch.value, sort_by: ['done']}}),
	{enabled: () => taskSearch.value !== ''},
)
const foundTasks = taskQuery.tasks
function findTasks(query: string) { taskSearch.value = query }


const canSubmit = computed(() =>
	// In edit mode the entry already has a valid container; an update that sends
	// neither keeps it, so don't block submit if the prefill lookup failed.
	isEditing.value || props.taskId !== undefined || selectedTask.value !== null || selectedProject.value !== null,
)

function smartFill() {
	from.value = smartFillStart(
		props.recentEntries,
		authStore.settings.frontendSettings.timeTrackingDefaultStart ?? '09:00',
		new Date(),
	)
	to.value = new Date()
}

// Whichever of task / project is set lands on the payload (XOR — enforced by canSubmit).
function applyTarget(payload: TimeEntryWritable) {
	if (props.taskId !== undefined) {
		payload.task_id = props.taskId
	} else if (selectedTask.value !== null) {
		payload.task_id = selectedTask.value.id
	} else if (selectedProject.value !== null) {
		payload.project_id = selectedProject.value.id
	}
}

function buildPayload(includeEnd: boolean): TimeEntryWritable {
	const payload: TimeEntryWritable = {
		comment: comment.value,
		start_time: (from.value ?? new Date()).toISOString(),
	}
	applyTarget(payload)
	// Saving a manual entry always has an end (an empty "To" means "until now");
	// only the Start-timer path omits it to create a running timer.
	if (includeEnd) {
		payload.end_time = (to.value ?? new Date()).toISOString()
	}
	return payload
}

function reset() {
	selectedTask.value = null
	selectedProject.value = null
	comment.value = ''
	from.value = new Date()
	to.value = null
}

// Prefill from the entry being edited; a null entry returns the form to create mode.
watch(() => props.entry, async (entry, _previous, onCleanup) => {
	let active = true
	onCleanup(() => { active = false })
	if (entry == null) {
		reset()
		return
	}
	comment.value = entry.comment
	from.value = parseDateOrNull(entry.start_time)
	to.value = parseDateOrNull(entry.end_time)
	// Bring the form into view — the edit button may be far down the list.
	await nextTick()
	if (!active) return
	formEl.value?.scrollIntoView({behavior: 'smooth', block: 'center'})
	if (props.taskId !== undefined) {
		return
	}
	if (entry.task_id > 0) {
		selectedProject.value = null
		selectedTask.value = null
		try {
			const task = await ensureTask(entry.task_id)
			if (active && selectedProject.value === null && selectedTask.value === null) {
				selectedTask.value = task
			}
		} catch {
			return
		}
	} else if (entry.project_id > 0) {
		selectedTask.value = null
		selectedProject.value = projectList.projects[entry.project_id] ?? null
	}
}, {immediate: true})

function draftIdentity() {
	return JSON.stringify([
		props.entry?.id,
		props.taskId,
		selectedTask.value?.id,
		selectedProject.value?.id,
		from.value,
		to.value,
		comment.value,
	])
}

async function submit(includeEnd: boolean) {
	if (!canSubmit.value || isSaving.value) {
		return
	}
	isSaving.value = true
	const draft = draftIdentity()
	try {
		const payload = buildPayload(includeEnd)
		// A started timer begins now (click time), not when the form first loaded.
		if (!includeEnd) {
			payload.start_time = new Date().toISOString()
		}
		await createMutation.mutateAsync(payload)
		if (draftIdentity() !== draft) return
		reset()
		emit('saved')
	} catch {
		return
	} finally {
		isSaving.value = false
	}
}

async function submitUpdate() {
	const entry = props.entry
	if (!canSubmit.value || isSaving.value || entry == null) {
		return
	}
	isSaving.value = true
	const draft = draftIdentity()
	try {
		const payload: UpdateTimeEntryInput = {
			id: entry.id,
			comment: comment.value,
			start_time: from.value?.toISOString() ?? entry.start_time,
			// A running entry stays running (null); a completed one can't be reopened,
			// so keep its end if "To" was cleared (the API rejects clearing it).
			end_time: to.value?.toISOString() ?? entry.end_time,
			task_id: 0,
			project_id: 0,
		}
		applyTarget(payload)
		await updateMutation.mutateAsync(payload)
		if (draftIdentity() !== draft) return
		emit('saved')
	} catch {
		return
	} finally {
		isSaving.value = false
	}
}

const saveEntry = () => (isEditing.value ? submitUpdate() : submit(true))
const startTimer = () => submit(false)
function cancelEdit() {
	emit('cancel')
}
</script>

<style lang="scss" scoped>
.field-columns {
	display: flex;
	gap: 1rem;

	> .field {
		flex: 1;
		min-inline-size: 0;
	}
}

.from-to-row {
	align-items: flex-end;
}

.smart-fill {
	display: inline-flex;
	align-items: center;
	justify-content: center;
	block-size: 2.5em;
	inline-size: 2.5em;
	border-radius: $radius;
	color: var(--primary);
	transition: background-color $transition;

	&:hover {
		background-color: var(--grey-100);
	}
}

.form-actions {
	display: flex;
	gap: .5rem;
}
</style>
