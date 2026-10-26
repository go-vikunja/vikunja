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
			v-if="items.length > 0"
			:padding="false"
			:has-content="false"
			:loading="isFetching"
		>
			<ul class="trashed-projects">
				<li
					v-for="project in items"
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
						<TrashRowActions
							:pending="isPending(project.id)"
							restore-cy="restoreProject"
							delete-forever-cy="deleteProjectForever"
							@restore="restore(project.id)"
							@deleteForever="confirmDeleteForever(project)"
						/>
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
			v-else-if="showEmptyState"
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
			:enabled="itemToDelete !== null"
			@close="cancelDeleteForever()"
			@submit="deleteForever()"
		>
			<template #header>
				<span>{{ $t('trash.deleteProjectForeverModal.header') }}</span>
			</template>
			<template #text>
				<p class="tw:text-balance">
					{{ $t('trash.deleteProjectForeverModal.text', {project: itemToDelete?.title}) }}
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
					{{ $t('trash.emptyProjectsModal.text', total) }}
				</p>
			</template>
		</Modal>
	</div>
</template>

<script setup lang="ts">
import {computed, reactive} from 'vue'
import {useQuery} from '@tanstack/vue-query'

import BaseButton from '@/components/base/BaseButton.vue'
import Card from '@/components/misc/Card.vue'
import ColorBubble from '@/components/misc/ColorBubble.vue'
import Modal from '@/components/misc/Modal.vue'
import PaginationEmit from '@/components/misc/PaginationEmit.vue'
import TrashRowActions from '@/components/trash/TrashRowActions.vue'

import {
	trashDaysLeft,
	trashedProjectsQuery,
	useDeleteProjectPermanentlyMutation,
	useEmptyProjectTrashMutation,
	useRestoreProjectMutation,
} from '@/client/queries/trash'
import {useDeletedLabel} from '@/composables/useDeletedLabel'
import {useTrashList} from '@/composables/useTrashList'
import {getHexColor} from '@/helpers/task'
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
	query: page => useQuery(computed(() => trashedProjectsQuery(page.value))),
	restore: useRestoreProjectMutation(),
	deleteForever: useDeleteProjectPermanentlyMutation(),
	empty: useEmptyProjectTrashMutation(),
})

const expanded = reactive(new Set<number>())
function toggle(projectId: number) {
	if (!expanded.delete(projectId)) {
		expanded.add(projectId)
	}
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
</style>
