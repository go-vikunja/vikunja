<template>
	<div>
		<Teleport
			defer
			to="#trash-actions"
		>
			<XButton
				v-cy="'emptyTaskTrash'"
				variant="secondary"
				icon="trash-alt"
				danger
				:aria-disabled="total === 0 || undefined"
				@click="openEmptyModal()"
			>
				{{ $t('trash.empty') }}
			</XButton>
		</Teleport>

		<Card
			v-if="tasks.length > 0"
			:padding="false"
			class="has-overflow"
			:has-content="false"
			:loading="isFetching"
		>
			<ul class="trashed-tasks">
				<li
					v-for="task in tasks"
					:key="task.id"
					v-cy="'trashedTask'"
					class="trashed-task"
				>
					<SingleTaskInlineReadonly
						:task="task"
						show-project
					/>
					<span
						v-tooltip="formatDateLong(task.deleted_at)"
						class="trashed-task-meta"
					>
						{{ deletedLabel(task) }} ·
						{{ $t('trash.daysLeftShort', trashDaysLeft(task.deleted_at ?? '')) }}
					</span>
					<span class="trashed-task-actions">
						<BaseButton
							v-tooltip="$t('trash.restore')"
							v-cy="'restoreTask'"
							class="trash-action"
							:aria-label="$t('trash.restore')"
							:aria-disabled="pendingRows.has(task.id) || undefined"
							@click="restore(task.id)"
						>
							<Icon icon="undo" />
						</BaseButton>
						<BaseButton
							v-tooltip="$t('trash.deleteForever')"
							v-cy="'deleteTaskForever'"
							class="trash-action trash-action--danger"
							:aria-label="$t('trash.deleteForever')"
							:aria-disabled="pendingRows.has(task.id) || undefined"
							@click="confirmDeleteForever(task)"
						>
							<Icon icon="trash-alt" />
						</BaseButton>
					</span>
				</li>
			</ul>
		</Card>
		<p
			v-else-if="isSuccess && !isFetching"
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

import BaseButton from '@/components/base/BaseButton.vue'
import Card from '@/components/misc/Card.vue'
import Modal from '@/components/misc/Modal.vue'
import PaginationEmit from '@/components/misc/PaginationEmit.vue'
import SingleTaskInlineReadonly from '@/components/tasks/partials/SingleTaskInlineReadonly.vue'

import {
	trashDaysLeft,
	trashedTasksQuery,
	useDeleteTaskPermanentlyMutation,
	useEmptyTaskTrashMutation,
	useRestoreTaskMutation,
} from '@/client/queries/trash'
import {clampPage} from '@/client/queries/pagination'
import type {TaskResponse} from '@/client/queries/tasks'
import {useDeletedLabel} from '@/composables/useDeletedLabel'
import {usePendingRows} from '@/composables/usePendingRows'
import {formatDateLong} from '@/helpers/time/formatDate'

const deletedLabel = useDeletedLabel({
	unknown: 'trash.deleted',
	by: 'trash.deletedBy',
	byYou: 'trash.deletedByYou',
})

const currentPage = ref(1)
const query = useQuery(computed(() => trashedTasksQuery(currentPage.value)))
const tasks = computed(() => query.data.value?.items ?? [])
const total = computed(() => query.data.value?.total ?? 0)
const totalPages = computed(() => query.data.value?.total_pages ?? 0)
const isFetching = query.isFetching
const isSuccess = query.isSuccess

// Restoring or deleting the last task of the last page leaves the page empty
watch(
	() => clampPage(currentPage.value, {
		data: query.data.value,
		isPlaceholderData: query.isPlaceholderData.value,
	}),
	page => {
		currentPage.value = page
	},
)

const restoreTask = useRestoreTaskMutation()
const deleteTaskPermanently = useDeleteTaskPermanentlyMutation()
const emptyTaskTrash = useEmptyTaskTrashMutation()

const taskToDelete = ref<TaskResponse | null>(null)
const showEmptyModal = ref(false)

const pendingRows = usePendingRows()

function restore(id: number) {
	return pendingRows.run(restoreTask, id)
}

function confirmDeleteForever(task: TaskResponse) {
	if (pendingRows.has(task.id)) {
		return
	}
	taskToDelete.value = task
}

function openEmptyModal() {
	if (total.value === 0) {
		return
	}
	showEmptyModal.value = true
}

async function deleteForever() {
	const task = taskToDelete.value
	taskToDelete.value = null
	if (!task) {
		return
	}
	await pendingRows.run(deleteTaskPermanently, task.id)
}

async function emptyTrash() {
	showEmptyModal.value = false
	try {
		await emptyTaskTrash.mutateAsync()
	} catch {
		return
	}
	currentPage.value = 1
}
</script>

<style lang="scss" scoped>
.trashed-tasks {
	list-style: none;
	margin: 0;
	padding: .5rem;
}

.trashed-task {
	display: flex;
	flex-wrap: wrap;
	align-items: center;
	padding: .4rem;
	border-radius: $radius;
	transition: background-color $transition;

	&:hover {
		background-color: var(--grey-100);
	}

	.task {
		cursor: default;
	}
}

.trashed-task-meta {
	margin-inline-start: auto;
	padding-inline: .75rem;
	color: var(--grey-500);
	font-size: .85rem;
	white-space: nowrap;
}

.trashed-task-actions {
	display: inline-flex;
	gap: .25rem;
}

.trash-action {
	display: inline-flex;
	align-items: center;
	justify-content: center;
	inline-size: 2rem;
	block-size: 2rem;
	border-radius: $radius;
	color: var(--grey-500);
	transition: color $transition, background-color $transition;

	&:hover {
		color: var(--primary);
		background-color: var(--grey-200);
	}

	&--danger:hover {
		color: var(--danger);
	}
}
</style>
