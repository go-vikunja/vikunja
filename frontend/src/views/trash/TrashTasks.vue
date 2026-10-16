<template>
	<div>
		<div class="trash-toolbar">
			<XButton
				v-cy="'emptyTaskTrash'"
				variant="secondary"
				icon="trash-alt"
				danger
				:disabled="total === 0"
				@click="showEmptyModal = true"
			>
				{{ $t('trash.empty') }}
			</XButton>
		</div>

		<Card
			v-if="tasks.length > 0"
			:padding="false"
			class="has-overflow"
			:has-content="false"
			:loading="isFetching"
		>
			<ul class="p-2 tasks">
				<li
					v-for="task in tasks"
					:key="task.id"
					v-cy="'trashedTask'"
				>
					<SingleTaskInProject
						:the-task="task"
						:can-mark-as-done="false"
						show-project
						readonly
					>
						<div class="trashed-task-meta">
							<span v-tooltip="formatDateLong(task.deleted_at)">
								{{ deletedLabel(task) }}
							</span>
							<span class="has-text-grey">
								{{ $t('trash.daysLeft', trashDaysLeft(task.deleted_at ?? '')) }}
							</span>
						</div>
						<div class="trashed-task-actions">
							<XButton
								v-cy="'restoreTask'"
								variant="tertiary"
								icon="undo"
								:loading="restoreTask.isPending.value && restoreTask.variables.value === task.id"
								@click="restoreTask.mutate(task.id)"
							>
								{{ $t('trash.restore') }}
							</XButton>
							<XButton
								v-cy="'deleteTaskForever'"
								variant="tertiary"
								icon="trash-alt"
								danger
								@click="taskToDelete = task"
							>
								{{ $t('trash.deleteForever') }}
							</XButton>
						</div>
					</SingleTaskInProject>
				</li>
			</ul>
		</Card>
		<p
			v-else-if="!isFetching"
			class="has-text-centered has-text-grey is-italic"
		>
			{{ $t('trash.noTasks') }}
		</p>

		<PaginationEmit
			v-if="totalPages > 1"
			:total-pages="totalPages"
			:current-page="currentPage"
			@pageChanged="currentPage = $event"
		/>

		<Modal
			:enabled="taskToDelete !== null"
			@close="taskToDelete = null"
			@submit="deleteForever()"
		>
			<template #header>
				<span>{{ $t('trash.deleteForeverModal.header') }}</span>
			</template>
			<template #text>
				<p class="tw:text-balance">
					{{ $t('trash.deleteForeverModal.text', {task: taskToDelete?.title}) }}
				</p>
			</template>
		</Modal>

		<Modal
			:enabled="showEmptyModal"
			@close="showEmptyModal = false"
			@submit="emptyTrash()"
		>
			<template #header>
				<span>{{ $t('trash.empty') }}</span>
			</template>
			<template #text>
				<p class="tw:text-balance">
					{{ $t('trash.emptyModal.text', total) }}
				</p>
			</template>
		</Modal>
	</div>
</template>

<script setup lang="ts">
import {computed, ref, watch} from 'vue'
import {useQuery} from '@tanstack/vue-query'
import {useI18n} from 'vue-i18n'

import Card from '@/components/misc/Card.vue'
import Modal from '@/components/misc/Modal.vue'
import PaginationEmit from '@/components/misc/PaginationEmit.vue'
import SingleTaskInProject from '@/components/tasks/partials/SingleTaskInProject.vue'

import {
	trashDaysLeft,
	trashedTasksQuery,
	useDeleteTaskPermanentlyMutation,
	useEmptyTaskTrashMutation,
	useRestoreTaskMutation,
} from '@/client/queries/trash'
import {clampPage} from '@/client/queries/pagination'
import type {TaskResponse} from '@/client/queries/tasks'
import {formatDateLong, formatDateSince} from '@/helpers/time/formatDate'
import {getDisplayName} from '@/helpers/user'

const {t} = useI18n({useScope: 'global'})

const currentPage = ref(1)
const query = useQuery(computed(() => trashedTasksQuery(currentPage.value)))
const tasks = computed(() => query.data.value?.items ?? [])
const total = computed(() => query.data.value?.total ?? 0)
const totalPages = computed(() => query.data.value?.total_pages ?? 0)
const isFetching = query.isFetching

// Restoring or deleting the last task of the last page leaves the page empty
watch(() => clampPage(currentPage.value, {data: query.data.value, isPlaceholderData: query.isPlaceholderData.value}), page => {
	currentPage.value = page
})

const restoreTask = useRestoreTaskMutation()
const deleteTaskPermanently = useDeleteTaskPermanentlyMutation()
const emptyTaskTrash = useEmptyTaskTrashMutation()

const taskToDelete = ref<TaskResponse | null>(null)
const showEmptyModal = ref(false)

function deletedLabel(task: TaskResponse) {
	const since = formatDateSince(task.deleted_at)
	return task.deleted_by
		? t('trash.deletedBy', {user: getDisplayName(task.deleted_by), since})
		: t('trash.deleted', {since})
}

async function deleteForever() {
	const task = taskToDelete.value
	taskToDelete.value = null
	if (task) {
		await deleteTaskPermanently.mutateAsync(task.id)
	}
}

async function emptyTrash() {
	showEmptyModal.value = false
	await emptyTaskTrash.mutateAsync()
	currentPage.value = 1
}
</script>

<style lang="scss" scoped>
.trash-toolbar {
	display: flex;
	justify-content: flex-end;
	margin-block-end: 1rem;
}

.trashed-task-meta {
	display: flex;
	flex-direction: column;
	align-items: flex-end;
	font-size: .85rem;
	margin-inline: auto .5rem;
}

.trashed-task-actions {
	display: flex;
	gap: .25rem;
}
</style>
