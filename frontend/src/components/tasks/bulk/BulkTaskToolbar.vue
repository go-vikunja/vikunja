<script setup lang="ts">
import {computed, ref, shallowReactive} from 'vue'
import {useI18n} from 'vue-i18n'

import Modal from '@/components/misc/Modal.vue'
import DatepickerInline from '@/components/input/DatepickerInline.vue'
import Multiselect from '@/components/input/Multiselect.vue'
import PrioritySelect from '@/components/tasks/partials/PrioritySelect.vue'
import PercentDoneSelect from '@/components/tasks/partials/PercentDoneSelect.vue'
import ProjectSearch from '@/components/tasks/partials/ProjectSearch.vue'
import Reminders from '@/components/tasks/partials/Reminders.vue'
import RepeatAfter from '@/components/tasks/partials/RepeatAfter.vue'
import User from '@/components/misc/User.vue'
import AssigneeList from '@/components/tasks/partials/AssigneeList.vue'
import Labels from '@/components/tasks/partials/Labels.vue'
import QuickAddMagic from '@/components/tasks/partials/QuickAddMagic.vue'
import BaseButton from '@/components/base/BaseButton.vue'

import TaskBulkService from '@/services/taskBulk'
import ProjectUserService from '@/services/projectUsers'
import TaskService from '@/services/task'
import {useBulkTaskSelection} from '@/stores/bulkTaskSelection'
import {useAuthStore} from '@/stores/auth'
import {useProjectStore} from '@/stores/projects'
import {useLabelStore} from '@/stores/labels'
import {useLabelStyles} from '@/composables/useLabelStyles'
import {includesById} from '@/helpers/utils'
import {getDisplayName} from '@/models/user'
import {getProjectTitle} from '@/helpers/getProjectTitle'
import {success} from '@/message'
import {getRandomColorHex} from '@/helpers/color/randomColor'

import TaskModel from '@/models/task'
import LabelModel from '@/models/label'
import type {ITask} from '@/modelTypes/ITask'
import type {IProject} from '@/modelTypes/IProject'
import type {ILabel} from '@/modelTypes/ILabel'
import type {IUser} from '@/modelTypes/IUser'
import type {ITaskReminder} from '@/modelTypes/ITaskReminder'
import {RELATION_KINDS, type IRelationKind} from '@/types/IRelationKind'
import {TASK_REPEAT_MODES} from '@/types/IRepeatMode'

const props = defineProps<{
	tasks: ITask[],
	projectId: IProject['id'],
}>()

const emit = defineEmits<{
	'updated': []
}>()

type BulkModal =
	| 'color'
	| 'dueDate'
	| 'startDate'
	| 'endDate'
	| 'move'
	| 'labels'
	| 'assignees'
	| 'relations'
	| 'reminders'
	| 'repeat'
	| 'duplicate'
	| 'delete'
	| null

type BulkListMode = 'add' | 'remove' | 'replace'

const selection = useBulkTaskSelection()
const authStore = useAuthStore()
const projectStore = useProjectStore()
const labelStore = useLabelStore()
const {getLabelStyles} = useLabelStyles()
const {t} = useI18n({useScope: 'global'})

const taskBulkService = shallowReactive(new TaskBulkService())
const projectUserService = shallowReactive(new ProjectUserService())
const taskService = shallowReactive(new TaskService())

const activeModal = ref<BulkModal>(null)

const labelMode = ref<BulkListMode>('add')
const labelQuery = ref('')
const bulkLabels = ref<ILabel[]>([])

const assigneeMode = ref<BulkListMode>('add')
const bulkAssignees = ref<IUser[]>([])
const foundUsers = ref<IUser[]>([])

const bulkColor = ref('#1973ff')
const bulkDueDate = ref<Date | null>(null)
const bulkStartDate = ref<Date | null>(null)
const bulkEndDate = ref<Date | null>(null)
const bulkProject = ref<IProject | null>(null)
const bulkReminders = ref<ITaskReminder[]>([])
const bulkRepeatTask = ref<ITask>(createEmptyRepeatTask())

const relatedTask = ref<ITask | null>(null)
const relatedTaskKind = ref<IRelationKind>(authStore.settings.frontendSettings.defaultTaskRelationType as IRelationKind)
const foundRelationTasks = ref<ITask[]>([])

const selectedTasks = computed(() =>
	props.tasks.filter(task => selection.isSelected(task.id)),
)

const isVisible = computed(() => selection.selectedCount > 1)
const loading = computed(() => taskBulkService.loading)

const selectedCountText = computed(() => `${selection.selectedCount} selected`)

const allSelectedAreFavorite = computed(() =>
	selectedTasks.value.length > 0 &&
	selectedTasks.value.every(task => task.isFavorite),
)

const allSelectedAreSubscribed = computed(() =>
	selectedTasks.value.length > 0 &&
	selectedTasks.value.every(task => task.subscription !== null),
)

const foundLabels = computed(() =>
	labelStore.filterLabelsByQuery(bulkLabels.value, labelQuery.value),
)

const mappedFoundRelationTasks = computed(() =>
	foundRelationTasks.value.map(task => ({
		...task,
		differentProject: task.projectId !== props.projectId && projectStore.projects[task.projectId]
			? getProjectTitle(projectStore.projects[task.projectId])
			: null,
	})),
)

function createEmptyRepeatTask() {
	return new TaskModel({
		repeatAfter: {
			amount: 0,
			type: 'days',
		},
		repeatMode: TASK_REPEAT_MODES.REPEAT_MODE_DEFAULT,
	})
}

function getSafeSelectedTasks() {
	return selectedTasks.value.filter(task => task.id > 0)
}

function openModal(modal: BulkModal) {
	activeModal.value = modal

	if (modal === 'color') {
		bulkColor.value = selectedTasks.value[0]?.hexColor
			? `#${selectedTasks.value[0].hexColor.replace(/^#/, '')}`
			: '#1973ff'
	}

	if (modal === 'dueDate') {
		bulkDueDate.value = selectedTasks.value[0]?.dueDate ?? null
	}

	if (modal === 'startDate') {
		bulkStartDate.value = selectedTasks.value[0]?.startDate ?? null
	}

	if (modal === 'endDate') {
		bulkEndDate.value = selectedTasks.value[0]?.endDate ?? null
	}

	if (modal === 'labels') {
		labelMode.value = 'add'
		labelQuery.value = ''
		bulkLabels.value = []
	}

	if (modal === 'assignees') {
		assigneeMode.value = 'add'
		bulkAssignees.value = []
		foundUsers.value = []
	}

	if (modal === 'relations') {
		relatedTask.value = null
		relatedTaskKind.value = authStore.settings.frontendSettings.defaultTaskRelationType as IRelationKind
		foundRelationTasks.value = []
	}

	if (modal === 'reminders') {
		bulkReminders.value = []
	}

	if (modal === 'repeat') {
		bulkRepeatTask.value = createEmptyRepeatTask()
	}

	if (modal === 'move') {
		bulkProject.value = null
	}
}

function closeModal() {
	activeModal.value = null
}

async function runAndRefresh(action: () => Promise<unknown>, message?: string) {
	const tasks = getSafeSelectedTasks()

	if (!isVisible.value || tasks.length === 0) {
		return
	}

	await action()

	if (message) {
		success({message})
	}

	emit('updated')
	closeModal()
}

async function toggleFavorite() {
	const makeFavorite = !allSelectedAreFavorite.value

	await runAndRefresh(() =>
		taskBulkService.setFavorite(getSafeSelectedTasks(), makeFavorite),
	)
}

async function toggleSubscribe() {
	await runAndRefresh(() =>
		allSelectedAreSubscribed.value
			? taskBulkService.unsubscribe(getSafeSelectedTasks())
			: taskBulkService.subscribe(getSafeSelectedTasks()),
	)
}

async function applyPriority(priority: number) {
	await runAndRefresh(() =>
		taskBulkService.updateTasks(getSafeSelectedTasks(), {
			priority,
		}),
	)
}

async function applyProgress(percentDone: number) {
	await runAndRefresh(() =>
		taskBulkService.updateTasks(getSafeSelectedTasks(), {
			percentDone,
		}),
	)
}

async function applyColor() {
	await runAndRefresh(() =>
		taskBulkService.updateTasks(getSafeSelectedTasks(), {
			hexColor: bulkColor.value,
		}),
	)
}

async function applyDate(field: 'dueDate' | 'startDate' | 'endDate') {
	const valueMap = {
		dueDate: bulkDueDate.value,
		startDate: bulkStartDate.value,
		endDate: bulkEndDate.value,
	}

	await runAndRefresh(() =>
		taskBulkService.updateTasks(getSafeSelectedTasks(), {
			[field]: valueMap[field],
		}),
	)
}

async function clearDate(field: 'dueDate' | 'startDate' | 'endDate') {
	await runAndRefresh(() =>
		taskBulkService.updateTasks(getSafeSelectedTasks(), {
			[field]: null,
		}),
	)
}

async function applyMove() {
	if (bulkProject.value === null || bulkProject.value.id === 0) {
		return
	}

	await runAndRefresh(() =>
		taskBulkService.moveTasks(getSafeSelectedTasks(), bulkProject.value as IProject),
	)
}

async function applyLabels() {
	if (bulkLabels.value.length === 0 && labelMode.value !== 'replace') {
		return
	}

	await runAndRefresh(async () => {
		if (labelMode.value === 'replace') {
			await taskBulkService.replaceLabels(getSafeSelectedTasks(), bulkLabels.value)
			return
		}

		if (labelMode.value === 'add') {
			await taskBulkService.addLabels(getSafeSelectedTasks(), bulkLabels.value)
			return
		}

		await taskBulkService.removeLabels(getSafeSelectedTasks(), bulkLabels.value)
	})
}

async function applyAssignees() {
	if (bulkAssignees.value.length === 0 && assigneeMode.value !== 'replace') {
		return
	}

	await runAndRefresh(async () => {
		if (assigneeMode.value === 'replace') {
			await taskBulkService.replaceAssignees(getSafeSelectedTasks(), bulkAssignees.value)
			return
		}

		if (assigneeMode.value === 'add') {
			await taskBulkService.addAssignees(getSafeSelectedTasks(), bulkAssignees.value)
			return
		}

		await taskBulkService.removeAssignees(getSafeSelectedTasks(), bulkAssignees.value)
	})
}

async function applyRelation() {
	if (relatedTask.value === null || relatedTask.value.id === 0) {
		return
	}

	await runAndRefresh(() =>
		taskBulkService.addRelation(getSafeSelectedTasks(), relatedTask.value!.id, relatedTaskKind.value),
	)
}

async function applyReminders() {
	await runAndRefresh(() =>
		taskBulkService.updateTasks(getSafeSelectedTasks(), {
			reminders: bulkReminders.value,
		}),
	)
}

async function clearReminders() {
	await runAndRefresh(() =>
		taskBulkService.updateTasks(getSafeSelectedTasks(), {
			reminders: [],
		}),
	)
}

async function applyRepeat() {
	await runAndRefresh(() =>
		taskBulkService.updateTasks(getSafeSelectedTasks(), {
			repeatAfter: bulkRepeatTask.value.repeatAfter,
			repeatMode: bulkRepeatTask.value.repeatMode,
		}),
	)
}

async function clearRepeat() {
	await runAndRefresh(() =>
		taskBulkService.updateTasks(getSafeSelectedTasks(), {
			repeatAfter: {
				amount: 0,
				type: 'days',
			},
			repeatMode: TASK_REPEAT_MODES.REPEAT_MODE_DEFAULT,
		}),
	)
}

async function duplicateSelected() {
	await runAndRefresh(() =>
		taskBulkService.duplicate(getSafeSelectedTasks()),
		t('task.detail.duplicateSuccess'),
	)
}

async function deleteSelected() {
	await runAndRefresh(
		async () => {
			await taskBulkService.deleteTasks(getSafeSelectedTasks())
			selection.clear()
		},
		'Deleted selected tasks.',
	)
}

function findLabel(query: string) {
	labelQuery.value = query
}

function selectLabel(label: ILabel) {
	if (!includesById(bulkLabels.value, label.id)) {
		bulkLabels.value.push(label)
	}
}

function removeLabel(label: ILabel) {
	const index = bulkLabels.value.findIndex(({id}) => id === label.id)

	if (index !== -1) {
		bulkLabels.value.splice(index, 1)
	}
}

async function createAndSelectLabel(title: string) {
	const trimmedTitle = title.trim()

	if (trimmedTitle === '') {
		return
	}

	const existing = Object.values(labelStore.labels).find(label =>
		label.title.toLowerCase() === trimmedTitle.toLowerCase(),
	)

	if (existing) {
		selectLabel(existing)
		return
	}

	const newLabel = await labelStore.createLabel(new LabelModel({
		title: trimmedTitle,
		hexColor: getRandomColorHex(),
	}))

	selectLabel(newLabel)
}

async function findUsers(query: string) {
	const response = await projectUserService.getAll({projectId: props.projectId}, {s: query}) as IUser[]

	foundUsers.value = response
		.filter(({id}) => !includesById(bulkAssignees.value, id))
		.map(user => {
			user.name = getDisplayName(user)
			return user
		})
}

function selectAssignee(user: IUser) {
	if (!includesById(bulkAssignees.value, user.id)) {
		bulkAssignees.value.push(user)
	}
}

function removeAssignee(user: IUser) {
	const index = bulkAssignees.value.findIndex(({id}) => id === user.id)

	if (index !== -1) {
		bulkAssignees.value.splice(index, 1)
	}
}

async function findRelationTasks(query: string) {
	const result = await taskService.getAll({}, {
		s: query,
		sort_by: 'done',
	})

	foundRelationTasks.value = result as ITask[]
}
</script>

<template>
	<div
		v-if="isVisible"
		class="bulk-task-toolbar d-print-none"
	>
		<div class="bulk-task-toolbar__count">
			{{ selectedCountText }}
		</div>

		<XButton
			variant="secondary"
			:icon="allSelectedAreFavorite ? 'star' : ['far', 'star']"
			:disabled="loading"
			@click="toggleFavorite"
		>
			{{ allSelectedAreFavorite ? $t('task.detail.actions.unfavorite') : $t('task.detail.actions.favorite') }}
		</XButton>

		<XButton
			variant="secondary"
			:icon="allSelectedAreSubscribed ? ['far', 'bell-slash'] : 'bell'"
			:disabled="loading"
			@click="toggleSubscribe"
		>
			{{ allSelectedAreSubscribed ? $t('task.subscription.unsubscribe') : $t('task.subscription.subscribe') }}
		</XButton>

		<XButton
			variant="secondary"
			icon="tags"
			:disabled="loading"
			@click="openModal('labels')"
		>
			{{ $t('task.detail.actions.label') }}
		</XButton>

		<XButton
			variant="secondary"
			icon="exclamation-circle"
			:disabled="loading"
			@click="openModal(null)"
		>
			<span class="bulk-task-toolbar__inline-select">
				{{ $t('task.detail.actions.priority') }}
				<PrioritySelect
					:model-value="selectedTasks[0]?.priority ?? 0"
					:disabled="loading"
					@update:modelValue="applyPriority"
				/>
			</span>
		</XButton>

		<XButton
			variant="secondary"
			icon="percent"
			:disabled="loading"
			@click="openModal(null)"
		>
			<span class="bulk-task-toolbar__inline-select">
				{{ $t('task.detail.actions.percentDone') }}
				<PercentDoneSelect
					:model-value="selectedTasks[0]?.percentDone ?? 0"
					:disabled="loading"
					@update:modelValue="applyProgress"
				/>
			</span>
		</XButton>

		<XButton
			variant="secondary"
			icon="fill-drip"
			:icon-color="bulkColor"
			:disabled="loading"
			@click="openModal('color')"
		>
			{{ $t('task.detail.actions.color') }}
		</XButton>

		<XButton
			variant="secondary"
			icon="users"
			:disabled="loading"
			@click="openModal('assignees')"
		>
			{{ $t('task.detail.actions.assign') }}
		</XButton>

		<XButton
			variant="secondary"
			icon="calendar"
			:disabled="loading"
			@click="openModal('dueDate')"
		>
			{{ $t('task.attributes.dueDate') }}
		</XButton>

		<XButton
			variant="secondary"
			icon="play"
			:disabled="loading"
			@click="openModal('startDate')"
		>
			{{ $t('task.attributes.startDate') }}
		</XButton>

		<XButton
			variant="secondary"
			icon="stop"
			:disabled="loading"
			@click="openModal('endDate')"
		>
			{{ $t('task.attributes.endDate') }}
		</XButton>

		<XButton
			variant="secondary"
			:icon="['far', 'clock']"
			:disabled="loading"
			@click="openModal('reminders')"
		>
			{{ $t('task.attributes.reminders') }}
		</XButton>

		<XButton
			variant="secondary"
			icon="history"
			:disabled="loading"
			@click="openModal('repeat')"
		>
			{{ $t('task.attributes.repeat') }}
		</XButton>

		<XButton
			variant="secondary"
			icon="sitemap"
			:disabled="loading"
			@click="openModal('relations')"
		>
			{{ $t('task.detail.actions.relatedTasks') }}
		</XButton>

		<XButton
			variant="secondary"
			icon="list"
			:disabled="loading"
			@click="openModal('move')"
		>
			{{ $t('task.detail.actions.moveProject') }}
		</XButton>

		<XButton
			variant="secondary"
			icon="copy"
			:disabled="loading"
			@click="openModal('duplicate')"
		>
			{{ $t('task.detail.actions.duplicate') }}
		</XButton>

		<XButton
			variant="secondary"
			icon="trash-alt"
			class="has-text-danger"
			:disabled="loading"
			@click="openModal('delete')"
		>
			Delete
		</XButton>

		<XButton
			variant="secondary"
			icon="times"
			:disabled="loading"
			@click="selection.clear"
		>
			Clear
		</XButton>
	</div>

	<Modal
		:enabled="activeModal !== null"
		:overflow="true"
		:wide="true"
		@close="closeModal"
	>
		<div class="bulk-task-modal">
			<div class="modal-header">
				<span v-if="activeModal === 'color'">
					<Icon icon="fill-drip" />
					{{ $t('task.detail.actions.color') }}
				</span>
				<span v-else-if="activeModal === 'dueDate'">
					<Icon icon="calendar" />
					{{ $t('task.attributes.dueDate') }}
				</span>
				<span v-else-if="activeModal === 'startDate'">
					<Icon icon="play" />
					{{ $t('task.attributes.startDate') }}
				</span>
				<span v-else-if="activeModal === 'endDate'">
					<Icon icon="stop" />
					{{ $t('task.attributes.endDate') }}
				</span>
				<span v-else-if="activeModal === 'move'">
					<Icon icon="list" />
					{{ $t('task.detail.actions.moveProject') }}
				</span>
				<span v-else-if="activeModal === 'labels'">
					<Icon icon="tags" />
					{{ $t('task.detail.actions.label') }}
				</span>
				<span v-else-if="activeModal === 'assignees'">
					<Icon icon="users" />
					{{ $t('task.detail.actions.assign') }}
				</span>
				<span v-else-if="activeModal === 'relations'">
					<Icon icon="sitemap" />
					{{ $t('task.detail.actions.relatedTasks') }}
				</span>
				<span v-else-if="activeModal === 'reminders'">
					<Icon :icon="['far', 'clock']" />
					{{ $t('task.attributes.reminders') }}
				</span>
				<span v-else-if="activeModal === 'repeat'">
					<Icon icon="history" />
					{{ $t('task.attributes.repeat') }}
				</span>
				<span v-else-if="activeModal === 'duplicate'">
					<Icon icon="copy" />
					{{ $t('task.detail.actions.duplicate') }}
				</span>
				<span v-else-if="activeModal === 'delete'">
					<Icon icon="trash-alt" />
					Delete tasks
				</span>
			</div>

			<p class="bulk-task-modal__subtitle">
				{{ selectedCountText }}
			</p>

			<div
				v-if="activeModal === 'color'"
				class="bulk-task-modal__body"
			>
				<input
					v-model="bulkColor"
					type="color"
					class="input color-input"
				>
			</div>

			<div
				v-else-if="activeModal === 'dueDate'"
				class="bulk-task-modal__body"
			>
				<DatepickerInline v-model="bulkDueDate" />
			</div>

			<div
				v-else-if="activeModal === 'startDate'"
				class="bulk-task-modal__body"
			>
				<DatepickerInline v-model="bulkStartDate" />
			</div>

			<div
				v-else-if="activeModal === 'endDate'"
				class="bulk-task-modal__body"
			>
				<DatepickerInline v-model="bulkEndDate" />
			</div>

			<div
				v-else-if="activeModal === 'move'"
				class="bulk-task-modal__body"
			>
				<ProjectSearch v-model="bulkProject" />
			</div>

			<div
				v-else-if="activeModal === 'labels'"
				class="bulk-task-modal__body"
			>
				<div class="field">
					<label class="label">Mode</label>
					<div class="select is-fullwidth">
						<select v-model="labelMode">
							<option value="add">
								Add labels
							</option>
							<option value="remove">
								Remove labels
							</option>
							<option value="replace">
								Replace labels
							</option>
						</select>
					</div>
				</div>

				<Multiselect
					v-model="bulkLabels"
					:loading="labelStore.isLoading"
					:placeholder="$t('task.label.placeholder')"
					:multiple="true"
					:search-results="foundLabels"
					label="title"
					:creatable="true"
					:create-placeholder="$t('task.label.createPlaceholder')"
					:search-delay="10"
					:close-after-select="false"
					@search="findLabel"
					@select="selectLabel"
					@create="createAndSelectLabel"
				>
					<template #tag="{item: label}">
						<span
							:style="getLabelStyles(label)"
							class="tag"
						>
							<span>{{ label.title }}</span>
							<BaseButton
								class="delete is-small"
								@click="removeLabel(label)"
							/>
						</span>
					</template>
					<template #searchResult="{option}">
						<span
							v-if="typeof option === 'string'"
							class="tag search-result"
						>
							<span>{{ option }}</span>
						</span>
						<span
							v-else
							:style="getLabelStyles(option)"
							class="tag search-result"
						>
							<span>{{ option.title }}</span>
						</span>
					</template>
				</Multiselect>

				<div
					v-if="bulkLabels.length > 0"
					class="bulk-task-modal__preview"
				>
					<Labels :labels="bulkLabels" />
				</div>
			</div>

			<div
				v-else-if="activeModal === 'assignees'"
				class="bulk-task-modal__body"
			>
				<div class="field">
					<label class="label">Mode</label>
					<div class="select is-fullwidth">
						<select v-model="assigneeMode">
							<option value="add">
								Add assignees
							</option>
							<option value="remove">
								Remove assignees
							</option>
							<option value="replace">
								Replace assignees
							</option>
						</select>
					</div>
				</div>

				<Multiselect
					v-model="bulkAssignees"
					class="edit-assignees"
					:class="{'has-assignees': bulkAssignees.length > 0}"
					:loading="projectUserService.loading"
					:placeholder="$t('task.assignee.placeholder')"
					:multiple="true"
					:search-results="foundUsers"
					label="name"
					:select-placeholder="$t('task.assignee.selectPlaceholder')"
					:autocomplete-enabled="false"
					@search="findUsers"
					@select="selectAssignee"
				>
					<template #items="{items}">
						<AssigneeList
							:assignees="items"
							can-remove
							@remove="removeAssignee"
						/>
					</template>
					<template #searchResult="{option: user}">
						<User
							:avatar-size="24"
							:show-username="true"
							:user="user"
						/>
					</template>
				</Multiselect>
			</div>

			<div
				v-else-if="activeModal === 'relations'"
				class="bulk-task-modal__body"
			>
				<label class="label">
					{{ $t('task.relation.new') }}
				</label>

				<div class="field task-relation-search-field">
					<Multiselect
						v-model="relatedTask"
						:placeholder="$t('task.relation.searchPlaceholder')"
						:loading="taskService.loading"
						:search-results="mappedFoundRelationTasks"
						label="title"
						:creatable="false"
						@search="findRelationTasks"
					>
						<template #searchResult="{option: task}">
							<span
								class="search-result"
								:class="{'is-strikethrough': task.done}"
							>
								<span
									v-if="task.projectId !== projectId"
									class="different-project"
								>
									<span v-if="task.differentProject !== null">
										{{ task.differentProject }} >
									</span>
								</span>
								{{ task.title }}
							</span>
						</template>
					</Multiselect>
					<QuickAddMagic />
				</div>

				<div class="field has-addons mbe-4">
					<div class="control is-expanded">
						<div class="select is-fullwidth has-defaults">
							<select v-model="relatedTaskKind">
								<option
									v-for="rk in RELATION_KINDS"
									:key="`bulk_relation_${rk}`"
									:value="rk"
								>
									{{ $t(`task.relation.kinds.${rk}`, 1) }}
								</option>
							</select>
						</div>
					</div>
				</div>
			</div>

			<div
				v-else-if="activeModal === 'reminders'"
				class="bulk-task-modal__body"
			>
				<Reminders
					v-model="bulkReminders"
					:allow-absolute="true"
				/>
			</div>

			<div
				v-else-if="activeModal === 'repeat'"
				class="bulk-task-modal__body"
			>
				<RepeatAfter v-model="bulkRepeatTask" />
			</div>

			<div
				v-else-if="activeModal === 'duplicate'"
				class="bulk-task-modal__body"
			>
				<p>
					Duplicate {{ selection.selectedCount }} selected tasks?
				</p>
			</div>

			<div
				v-else-if="activeModal === 'delete'"
				class="bulk-task-modal__body"
			>
				<p>
					Delete {{ selection.selectedCount }} selected tasks?
				</p>
				<p class="has-text-danger">
					This cannot be undone.
				</p>
			</div>

			<div class="actions">
				<XButton
					variant="tertiary"
					class="has-text-danger"
					@click="closeModal"
				>
					{{ $t('misc.cancel') }}
				</XButton>

				<XButton
					v-if="activeModal === 'dueDate'"
					variant="secondary"
					:disabled="loading"
					@click="clearDate('dueDate')"
				>
					Clear
				</XButton>

				<XButton
					v-if="activeModal === 'startDate'"
					variant="secondary"
					:disabled="loading"
					@click="clearDate('startDate')"
				>
					Clear
				</XButton>

				<XButton
					v-if="activeModal === 'endDate'"
					variant="secondary"
					:disabled="loading"
					@click="clearDate('endDate')"
				>
					Clear
				</XButton>

				<XButton
					v-if="activeModal === 'reminders'"
					variant="secondary"
					:disabled="loading"
					@click="clearReminders"
				>
					Clear
				</XButton>

				<XButton
					v-if="activeModal === 'repeat'"
					variant="secondary"
					:disabled="loading"
					@click="clearRepeat"
				>
					Clear
				</XButton>

				<XButton
					v-cy="'modalPrimary'"
					variant="primary"
					:shadow="false"
					:loading="loading"
					@click="
						activeModal === 'color' ? applyColor() :
						activeModal === 'dueDate' ? applyDate('dueDate') :
						activeModal === 'startDate' ? applyDate('startDate') :
						activeModal === 'endDate' ? applyDate('endDate') :
						activeModal === 'move' ? applyMove() :
						activeModal === 'labels' ? applyLabels() :
						activeModal === 'assignees' ? applyAssignees() :
						activeModal === 'relations' ? applyRelation() :
						activeModal === 'reminders' ? applyReminders() :
						activeModal === 'repeat' ? applyRepeat() :
						activeModal === 'duplicate' ? duplicateSelected() :
						activeModal === 'delete' ? deleteSelected() :
						undefined
					"
				>
					{{ $t('misc.doit') }}
				</XButton>
			</div>
		</div>
	</Modal>
</template>

<style scoped lang="scss">
.bulk-task-toolbar {
	display: flex;
	flex-wrap: wrap;
	align-items: center;
	gap: .5rem;
	padding: .75rem;
	margin-block: .75rem;
	border: 1px solid var(--grey-200);
	border-radius: var(--radius);
	background: var(--white);
	box-shadow: var(--shadow-sm);
}

.bulk-task-toolbar__count {
	font-weight: 600;
	color: var(--grey-700);
	margin-inline-end: .25rem;
	white-space: nowrap;
}

.bulk-task-toolbar__inline-select {
	display: inline-flex;
	align-items: center;
	gap: .5rem;
}

.bulk-task-toolbar__inline-select :deep(.select select) {
	block-size: 1.75rem;
	min-block-size: 1.75rem;
	padding-block: 0;
}

.bulk-task-modal {
	text-align: start;
	padding: 2rem;
}

.bulk-task-modal .modal-header {
	display: flex;
	align-items: center;
	gap: .5rem;
	font-size: 1.5rem;
	font-weight: 700;
	margin-block-end: .25rem;
}

.bulk-task-modal .modal-header span {
	display: inline-flex;
	align-items: center;
	gap: .5rem;
}

.bulk-task-modal__subtitle {
	color: var(--grey-500);
	margin-block-end: 1.5rem;
}

.bulk-task-modal__body {
	margin-block-end: 1.5rem;
}

.bulk-task-modal__preview {
	margin-block-start: .75rem;
}

.bulk-task-modal .actions {
	display: flex;
	justify-content: flex-end;
	align-items: center;
	gap: .5rem;
	margin-block-start: 1.5rem;
}

.color-input {
	inline-size: 100%;
	min-block-size: 3rem;
	padding: .25rem;
}

.edit-assignees.has-assignees.multiselect :deep(.input) {
	padding-inline-start: 0;
}

.task-relation-search-field {
	position: relative;
}

.different-project {
	color: var(--grey-500);
}

.is-strikethrough {
	text-decoration: line-through;
}

.tag {
	margin: .25rem !important;
}

.tag.search-result {
	margin: 0 !important;
}
</style>