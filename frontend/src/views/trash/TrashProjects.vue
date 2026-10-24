<template>
	<div>
		<Teleport
			defer
			to="#trash-actions"
		>
			<XButton
				v-cy="'emptyProjectTrash'"
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
				>
					<div class="trashed-project">
						<ColorBubble
							v-if="getHexColor(project.hex_color)"
							:color="getHexColor(project.hex_color)"
						/>
						<span class="trashed-project__title">{{ project.title }}</span>
						<BaseButton
							v-if="project.trashed_sub_projects?.length"
							v-cy="'toggleTrashedSubProjects'"
							class="trashed-project__toggle"
							:aria-expanded="expanded.has(project.id)"
							@click="toggle(project.id)"
						>
							<Icon
								icon="chevron-down"
								class="trashed-project__chevron"
								:class="{'is-collapsed': !expanded.has(project.id)}"
							/>
							{{ $t('trash.subProjects', project.trashed_sub_projects.length) }}
						</BaseButton>
						<span
							v-tooltip="formatDateLong(project.deleted_at)"
							class="trashed-project__meta"
						>
							{{ deletedLabel(project) }} · {{ $t('trash.daysLeftShort', trashDaysLeft(project.deleted_at ?? '')) }}
						</span>
						<span class="trashed-project__actions">
							<BaseButton
								v-tooltip="$t('trash.restore')"
								v-cy="'restoreProject'"
								class="trash-action"
								:aria-label="$t('trash.restore')"
								:aria-disabled="isRestoring(project.id) || undefined"
								@click="restore(project.id)"
							>
								<Icon icon="undo" />
							</BaseButton>
							<BaseButton
								v-tooltip="$t('trash.deleteForever')"
								v-cy="'deleteProjectForever'"
								class="trash-action trash-action--danger"
								:aria-label="$t('trash.deleteForever')"
								@click="projectToDelete = project"
							>
								<Icon icon="trash-alt" />
							</BaseButton>
						</span>
					</div>
					<ul
						v-if="expanded.has(project.id)"
						v-cy="'trashedSubProjects'"
						class="trashed-sub-projects"
					>
						<li
							v-for="sub in project.trashed_sub_projects"
							:key="sub.id"
							:style="{paddingInlineStart: `${(sub.depth ?? 1) * 1.25}rem`}"
						>
							{{ sub.title }}
						</li>
					</ul>
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
import {computed, reactive, ref, watch} from 'vue'
import {useQuery} from '@tanstack/vue-query'

import BaseButton from '@/components/base/BaseButton.vue'
import Card from '@/components/misc/Card.vue'
import ColorBubble from '@/components/misc/ColorBubble.vue'
import Modal from '@/components/misc/Modal.vue'
import PaginationEmit from '@/components/misc/PaginationEmit.vue'

import type {ProjectResponse} from '@/client/queries/projects'
import {
	trashDaysLeft,
	trashedProjectsQuery,
	useDeleteProjectPermanentlyMutation,
	useEmptyProjectTrashMutation,
	useRestoreProjectMutation,
} from '@/client/queries/trash'
import {clampPage} from '@/client/queries/pagination'
import {useDeletedLabel} from '@/composables/useDeletedLabel'
import {getHexColor} from '@/helpers/task'
import {formatDateLong} from '@/helpers/time/formatDate'

const deletedLabel = useDeletedLabel('trash.deleted')

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

const projectToDelete = ref<ProjectResponse | null>(null)
const showEmptyModal = ref(false)

const expanded = reactive(new Set<number>())
function toggle(projectId: number) {
	if (!expanded.delete(projectId)) {
		expanded.add(projectId)
	}
}

function isRestoring(projectId: number) {
	return restoreProject.isPending.value && restoreProject.variables.value === projectId
}

function restore(projectId: number) {
	if (isRestoring(projectId)) return
	restoreProject.mutate(projectId)
}

function openEmptyModal() {
	if (total.value === 0) return
	showEmptyModal.value = true
}

async function deleteForever() {
	const project = projectToDelete.value
	projectToDelete.value = null
	if (project) {
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
.trashed-projects {
	list-style: none;
	margin: 0;
	padding: .5rem;
}

.trashed-project {
	display: flex;
	flex-wrap: wrap;
	align-items: center;
	gap: .5rem;
	padding: .4rem;
	border-radius: $radius;

	&:hover {
		background-color: var(--grey-100);
	}
}

.trashed-project__title {
	min-inline-size: 0;
	overflow-wrap: anywhere;
}

.trashed-project__toggle {
	display: inline-flex;
	align-items: center;
	gap: .25rem;
	color: var(--grey-500);
	font-size: .9rem;

	&:hover {
		color: var(--primary);
	}
}

.trashed-project__chevron {
	font-size: .7rem;
	transition: transform $transition;

	&.is-collapsed {
		transform: rotate(-90deg);
	}
}

.trashed-project__meta {
	margin-inline-start: auto;
	padding-inline: .25rem .5rem;
	color: var(--grey-500);
	font-size: .85rem;
	white-space: nowrap;
}

.trashed-project__actions {
	display: inline-flex;
	gap: .25rem;
}

.trashed-sub-projects {
	list-style: none;
	margin: 0 0 .5rem;
	padding: 0 .4rem;
	color: var(--grey-600);
	font-size: .9rem;

	li {
		padding-block: .15rem;
	}
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
