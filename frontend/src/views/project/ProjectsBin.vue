<template>
	<div
		class="content loader-container"
		:class="{'is-loading': isLoading}"
	>
		<h1>{{ $t('project.bin.title') }}</h1>

		<p v-if="deletedProjects.length === 0 && !isLoading">
			{{ $t('project.bin.empty') }}
		</p>

		<div
			v-for="project in deletedProjects"
			:key="project.id"
			class="deleted-project"
		>
			<div class="deleted-project-info">
				<span class="deleted-project-title">{{ project.title }}</span>
				<span class="deleted-project-meta">
					{{ $t('project.bin.deletedOn', {date: formatDateShort(project.deletedAt)}) }}
					&mdash;
					{{ $t('project.bin.daysRemaining', {days: daysRemaining(project.deletedAt)}) }}
				</span>
			</div>
			<XButton
				variant="secondary"
				:loading="restoring === project.id"
				@click="restoreProject(project)"
			>
				{{ $t('project.bin.restore') }}
			</XButton>
		</div>
	</div>
</template>

<script setup lang="ts">
import {ref, onMounted} from 'vue'
import {useI18n} from 'vue-i18n'
import {useTitle} from '@/composables/useTitle'
import {success} from '@/message'
import {useProjectStore} from '@/stores/projects'
import {formatDateShort} from '@/helpers/time/formatDate'

import type {IProject} from '@/modelTypes/IProject'

const SOFT_DELETE_RETENTION_DAYS = 30

const {t} = useI18n({useScope: 'global'})
const projectStore = useProjectStore()

useTitle(() => t('project.bin.title'))

const deletedProjects = ref<IProject[]>([])
const isLoading = ref(false)
const restoring = ref<number | null>(null)

onMounted(async () => {
	isLoading.value = true
	try {
		deletedProjects.value = await projectStore.fetchDeletedProjects()
	} finally {
		isLoading.value = false
	}
})

function daysRemaining(deletedAt: Date | null): number {
	if (!deletedAt) return 0
	const deleted = new Date(deletedAt)
	const purgeDate = new Date(deleted.getTime() + SOFT_DELETE_RETENTION_DAYS * 24 * 60 * 60 * 1000)
	const remaining = Math.ceil((purgeDate.getTime() - Date.now()) / (24 * 60 * 60 * 1000))
	return Math.max(0, remaining)
}

async function restoreProject(project: IProject) {
	restoring.value = project.id
	try {
		await projectStore.restoreProject(project.id)
		deletedProjects.value = deletedProjects.value.filter(p => p.id !== project.id)
		success({message: t('project.bin.restoreSuccess')})
	} finally {
		restoring.value = null
	}
}
</script>

<style lang="scss" scoped>
.deleted-project {
	display: flex;
	justify-content: space-between;
	align-items: center;
	padding: 1rem;
	border-block-end: 1px solid var(--grey-200);

	&:last-child {
		border-block-end: none;
	}
}

.deleted-project-info {
	display: flex;
	flex-direction: column;
	gap: 0.25rem;
}

.deleted-project-title {
	font-weight: bold;
}

.deleted-project-meta {
	font-size: 0.875rem;
	color: var(--grey-500);
}
</style>
