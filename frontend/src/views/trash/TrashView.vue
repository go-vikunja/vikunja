<template>
	<div
		class="content loader-container is-max-width-desktop"
		:class="{'is-loading': trashService.loading}"
	>
		<div class="trash-header">
			<div>
				<h1>{{ $t('trash.title') }}</h1>
				<p class="has-text-grey">
					{{ $t('trash.subtitle') }}
				</p>
			</div>
			<XButton
				v-if="tasks.length > 0"
				danger
				icon="trash-alt"
				@click="showEmptyConfirm = true"
			>
				{{ $t('trash.empty') }}
			</XButton>
		</div>

		<p
			v-if="tasks.length === 0 && !trashService.loading"
			class="has-text-centered has-text-grey is-italic"
		>
			{{ $t('trash.noItems') }}
		</p>

		<div
			v-for="task in tasks"
			:key="task.id"
			class="trash-item box"
		>
			<div class="trash-item-info">
				<strong>{{ task.title }}</strong>
				<div class="trash-item-meta has-text-grey">
					<span>{{ $t('trash.deletedAgo', {days: getDaysAgo(task.deletedAt)}) }}</span>
					<span class="separator">|</span>
					<span>{{ $t('trash.daysRemaining', {days: getDaysRemaining(task.deletedAt)}) }}</span>
				</div>
			</div>
			<div class="trash-item-actions">
				<XButton
					icon="undo"
					@click="restoreTask(task.id)"
				>
					{{ $t('trash.restore') }}
				</XButton>
				<XButton
					danger
					icon="trash-alt"
					@click="confirmDeletePermanently(task.id)"
				>
					{{ $t('trash.deletePermanently') }}
				</XButton>
			</div>
		</div>

		<Modal
			:enabled="showEmptyConfirm"
			@close="showEmptyConfirm = false"
			@submit="emptyTrash()"
		>
			<template #header>
				<span>{{ $t('trash.empty') }}</span>
			</template>

			<template #text>
				<p>{{ $t('trash.emptyConfirm') }}</p>
			</template>
		</Modal>

		<Modal
			:enabled="showDeleteConfirm"
			@close="showDeleteConfirm = false"
			@submit="deletePermanently()"
		>
			<template #header>
				<span>{{ $t('trash.deletePermanently') }}</span>
			</template>

			<template #text>
				<p>{{ $t('trash.deletePermanentlyConfirm') }}</p>
			</template>
		</Modal>
	</div>
</template>

<script setup lang="ts">
import {ref, shallowReactive} from 'vue'
import {useI18n} from 'vue-i18n'

import TrashService from '@/services/trash'
import {success} from '@/message'
import {useTitle} from '@/composables/useTitle'
import type {ITask} from '@/modelTypes/ITask'

const {t} = useI18n({useScope: 'global'})
useTitle(() => t('trash.title'))

const tasks = ref<ITask[]>([])
const trashService = shallowReactive(new TrashService())
const showEmptyConfirm = ref(false)
const showDeleteConfirm = ref(false)
const taskToDelete = ref<number | null>(null)

async function loadTasks() {
	tasks.value = await trashService.getAll()
}

loadTasks()

function getDaysAgo(deletedAt: Date | null): number {
	if (!deletedAt) {
		return 0
	}
	const now = new Date()
	const deleted = new Date(deletedAt)
	return Math.floor((now.getTime() - deleted.getTime()) / (1000 * 60 * 60 * 24))
}

function getDaysRemaining(deletedAt: Date | null): number {
	if (!deletedAt) {
		return 30
	}
	const daysAgo = getDaysAgo(deletedAt)
	return Math.max(0, 30 - daysAgo)
}

async function restoreTask(taskId: number) {
	await trashService.restore(taskId)
	success({message: t('trash.restoreSuccess')})
	await loadTasks()
}

function confirmDeletePermanently(taskId: number) {
	taskToDelete.value = taskId
	showDeleteConfirm.value = true
}

async function deletePermanently() {
	if (taskToDelete.value === null) {
		return
	}
	await trashService.deletePermanently(taskToDelete.value)
	success({message: t('trash.deletePermanentlySuccess')})
	showDeleteConfirm.value = false
	taskToDelete.value = null
	await loadTasks()
}

async function emptyTrash() {
	await trashService.emptyTrash()
	success({message: t('trash.emptySuccess')})
	showEmptyConfirm.value = false
	tasks.value = []
}
</script>

<style lang="scss" scoped>
.trash-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-block-end: 1rem;
}

.trash-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem;
}

.trash-item-info {
  flex: 1;
}

.trash-item-meta {
  font-size: 0.85rem;
  margin-block-start: 0.25rem;

  .separator {
    margin-inline: 0.5rem;
  }
}

.trash-item-actions {
  display: flex;
  gap: 0.5rem;
  flex-shrink: 0;
}
</style>
