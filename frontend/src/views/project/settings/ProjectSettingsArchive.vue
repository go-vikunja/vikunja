<template>
	<Modal
		@close="$router.back()"
		@submit="archiveProject()"
	>
		<template #header>
			<span>{{ project.isArchived ? $t('project.archive.unarchive') : $t('project.archive.archive') }}</span>
		</template>
		
		<template #text>
			<p>{{ project.isArchived ? $t('project.archive.unarchiveText') : $t('project.archive.archiveText') }}</p>
		</template>
	</Modal>
</template>

<script setup lang="ts">
import {computed} from 'vue'
import {useRouter, useRoute} from 'vue-router'
import {useI18n} from 'vue-i18n'

import {success} from '@/message'
import {useTitle} from '@/composables/useTitle'

import {useBaseStore} from '@/stores/base'
import {useProjectStore} from '@/stores/projects'

defineOptions({name: 'ProjectSettingArchive'})

const {t} = useI18n({useScope: 'global'})
const projectStore = useProjectStore()
const router = useRouter()
const route = useRoute()

const project = computed(() => projectStore.projects[route.params.projectId])
useTitle(() => t('project.archive.title', {project: project.value.title}))

async function archiveProject() {
	try {
		const newProject = await projectStore.updateProject({
			...project.value,
			isArchived: !project.value.isArchived,
		})
		useBaseStore().setCurrentProject(newProject)
		success({message: t('project.archive.success')})
		await projectStore.loadAllProjects()
	} finally {
		router.back()
	}
}
</script>
