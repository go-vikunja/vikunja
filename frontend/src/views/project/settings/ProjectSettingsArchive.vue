<template>
	<Modal
		@close="$router.back()"
		@submit="archiveProject()"
	>
		<template #header>
			<span>{{ project?.is_archived ? $t('project.archive.unarchive') : $t('project.archive.archive') }}</span>
		</template>
		
		<template #text>
			<p>{{ project?.is_archived ? $t('project.archive.unarchiveText') : $t('project.archive.archiveText') }}</p>
		</template>
	</Modal>
</template>

<script setup lang="ts">
import {computed} from 'vue'
import {useRouter, useRoute} from 'vue-router'
import {useI18n} from 'vue-i18n'

import {useUpdateProjectMutation} from '@/client/queries/projects'
import {useTitle} from '@/composables/useTitle'

import {useBaseStore} from '@/stores/base'
import {useProjects} from '@/composables/useProjects'

defineOptions({name: 'ProjectSettingArchive'})

const {t} = useI18n({useScope: 'global'})
const projectList = useProjects()
const updateMutation = useUpdateProjectMutation(t('project.archive.success'))
const router = useRouter()
const route = useRoute()

const project = computed(() => projectList.projects[route.params.projectId])
useTitle(() => project.value?.title ? t('project.archive.title', {project: project.value.title}) : '')

async function archiveProject() {
	if (!project.value) {
		return
	}

	try {
		const newProject = await updateMutation.mutateAsync({
			...project.value,
			is_archived: !project.value.is_archived,
		})
		useBaseStore().setCurrentProject(newProject)
	} finally {
		router.back()
	}
}
</script>
