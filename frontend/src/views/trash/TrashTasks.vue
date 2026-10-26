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

		<div
			ref="listElement"
			tabindex="-1"
		>
			<Card
				v-if="items.length > 0"
				:padding="false"
				class="has-overflow"
				:has-content="false"
				:loading="isFetching"
			>
				<ul class="trashed-tasks">
					<li
						v-for="task in items"
						:key="task.id"
						v-cy="'trashedTask'"
						:data-trash-row="task.id"
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
						<TrashRowActions
							:pending="isPending(task.id)"
							restore-cy="restoreTask"
							delete-forever-cy="deleteTaskForever"
							@restore="restore(task.id)"
							@deleteForever="confirmDeleteForever(task)"
						/>
					</li>
				</ul>
			</Card>
			<p
				v-else-if="showEmptyState"
				ref="emptyStateElement"
				tabindex="-1"
				class="has-text-centered has-text-grey is-italic"
			>
				{{ $t('trash.noTasks') }}
			</p>
		</div>

		<PaginationEmit
			v-if="totalPages > 1"
			:total-pages="totalPages"
			:current-page="currentPage"
			@pageChanged="currentPage = $event"
		/>

		<Modal
			:enabled="itemToDelete !== null"
			@close="cancelDeleteForever()"
			@submit="deleteForever()"
		>
			<template #header>
				<span>{{ $t('trash.deleteForeverModal.header') }}</span>
			</template>
			<template #text>
				<p class="tw:text-balance">
					{{ $t('trash.deleteForeverModal.text', {task: itemToDelete?.title}) }}
				</p>
			</template>
		</Modal>

		<Modal
			:enabled="showEmptyModal"
			@close="closeEmptyModal()"
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
import {computed} from 'vue'
import {useQuery} from '@tanstack/vue-query'

import Card from '@/components/misc/Card.vue'
import Modal from '@/components/misc/Modal.vue'
import PaginationEmit from '@/components/misc/PaginationEmit.vue'
import SingleTaskInlineReadonly from '@/components/tasks/partials/SingleTaskInlineReadonly.vue'
import TrashRowActions from '@/components/trash/TrashRowActions.vue'

import {
	trashDaysLeft,
	trashedTasksQuery,
	useDeleteTaskPermanentlyMutation,
	useEmptyTaskTrashMutation,
	useRestoreTaskMutation,
} from '@/client/queries/trash'
import {useDeletedLabel} from '@/composables/useDeletedLabel'
import {useTrashList} from '@/composables/useTrashList'
import {formatDateLong} from '@/helpers/time/formatDate'

const deletedLabel = useDeletedLabel({
	unknown: 'trash.deleted',
	by: 'trash.deletedBy',
	byYou: 'trash.deletedByYou',
})

const {
	currentPage,
	items,
	total,
	totalPages,
	isFetching,
	showEmptyState,
	listElement,
	emptyStateElement,
	isPending,
	itemToDelete,
	showEmptyModal,
	restore,
	confirmDeleteForever,
	cancelDeleteForever,
	deleteForever,
	openEmptyModal,
	closeEmptyModal,
	emptyTrash,
} = useTrashList({
	query: page => useQuery(computed(() => trashedTasksQuery(page.value))),
	restore: useRestoreTaskMutation(),
	deleteForever: useDeleteTaskPermanentlyMutation(),
	empty: useEmptyTaskTrashMutation(),
})
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
</style>
