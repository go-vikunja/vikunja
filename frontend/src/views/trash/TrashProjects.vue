<template>
	<div>
		<div class="trash-toolbar">
			<XButton
				v-cy="'emptyProjectTrash'"
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
			v-if="projects.length > 0"
			:padding="false"
			:has-content="false"
			:loading="isFetching"
		>
			<ul class="trashed-projects">
				<li
					v-for="project in projects"
					:key="project.id"
					v-cy="'trashedProject'"
					class="trashed-project"
				>
					<div class="trashed-project__title">
						<span class="has-text-weight-semibold">{{ project.title }}</span>
						<span
							v-if="project.trashed_sub_projects"
							class="has-text-grey mis-2"
						>
							{{ $t('trash.subProjects', project.trashed_sub_projects) }}
						</span>
					</div>
					<div class="trashed-project__meta">
						<span v-tooltip="formatDateLong(project.deleted_at)">
							{{ deletedLabel(project) }}
						</span>
						<span class="has-text-grey">
							{{ $t('trash.daysLeft', trashDaysLeft(project.deleted_at ?? '')) }}
						</span>
					</div>
					<div class="trashed-project__actions">
						<XButton
							v-cy="'restoreProject'"
							variant="tertiary"
							icon="undo"
							:loading="restoreProject.isPending.value && restoreProject.variables.value === project.id"
							@click="restoreProject.mutate(project.id!)"
						>
							{{ $t('trash.restore') }}
						</XButton>
						<XButton
							v-cy="'deleteProjectForever'"
							variant="tertiary"
							icon="trash-alt"
							danger
							@click="projectToDelete = project"
						>
							{{ $t('trash.deleteForever') }}
						</XButton>
					</div>
				</li>
			</ul>
		</Card>
		<p
			v-else-if="!isFetching"
			class="has-text-centered has-text-grey is-italic"
		>
			{{ $t('trash.noProjects') }}
		</p>

		<PaginationEmit
			v-if="totalPages > 1"
			:total-pages="totalPages"
			:current-page="currentPage"
			@pageChanged="currentPage = $event"
		/>

		<Modal
			:enabled="projectToDelete !== null"
			@close="projectToDelete = null"
			@submit="deleteForever()"
		>
			<template #header>
				<span>{{ $t('trash.deleteProjectForeverModal.header') }}</span>
			</template>
			<template #text>
				<p class="tw:text-balance">
					{{ $t('trash.deleteProjectForeverModal.text', {project: projectToDelete?.title}) }}
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
					{{ $t('trash.emptyProjectsModal.text', total) }}
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

import type {Project} from '@/client/generated'
import {
	trashDaysLeft,
	trashedProjectsQuery,
	useDeleteProjectPermanentlyMutation,
	useEmptyProjectTrashMutation,
	useRestoreProjectMutation,
} from '@/client/queries/trash'
import {clampPage} from '@/client/queries/pagination'
import {formatDateLong, formatDateSince} from '@/helpers/time/formatDate'
import {getDisplayName} from '@/helpers/user'

const {t} = useI18n({useScope: 'global'})

const currentPage = ref(1)
const query = useQuery(computed(() => trashedProjectsQuery(currentPage.value)))
const projects = computed(() => query.data.value?.items ?? [])
const total = computed(() => query.data.value?.total ?? 0)
const totalPages = computed(() => query.data.value?.total_pages ?? 0)
const isFetching = query.isFetching

// Restoring or deleting the last project of the last page leaves the page empty
watch(() => clampPage(currentPage.value, {data: query.data.value, isPlaceholderData: query.isPlaceholderData.value}), page => {
	currentPage.value = page
})

const restoreProject = useRestoreProjectMutation()
const deleteProjectPermanently = useDeleteProjectPermanentlyMutation()
const emptyProjectTrash = useEmptyProjectTrashMutation()

const projectToDelete = ref<Project | null>(null)
const showEmptyModal = ref(false)

function deletedLabel(project: Project) {
	const since = formatDateSince(project.deleted_at)
	return project.deleted_by
		? t('trash.deletedBy', {user: getDisplayName(project.deleted_by), since})
		: t('trash.deleted', {since})
}

async function deleteForever() {
	const project = projectToDelete.value
	projectToDelete.value = null
	if (project?.id) {
		await deleteProjectPermanently.mutateAsync(project.id)
	}
}

async function emptyTrash() {
	showEmptyModal.value = false
	await emptyProjectTrash.mutateAsync()
	currentPage.value = 1
}
</script>

<style lang="scss" scoped>
.trash-toolbar {
	display: flex;
	justify-content: flex-end;
	margin-block-end: 1rem;
}

.trashed-project {
	display: flex;
	align-items: center;
	gap: 1rem;
	padding: .75rem 1rem;

	& + & {
		border-block-start: 1px solid var(--grey-200);
	}
}

.trashed-project__title {
	flex: 1;
	min-inline-size: 0;
}

.trashed-project__meta {
	display: flex;
	flex-direction: column;
	align-items: flex-end;
	font-size: .85rem;
}

.trashed-project__actions {
	display: flex;
	gap: .25rem;
}
</style>
