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

<script lang="ts">
export default {name: 'ProjectSettingArchive'}
</script>

<script setup lang="ts">
import {computed} from 'vue'
import {useRouter, type RouteLocationNormalizedLoaded} from 'vue-router'
import {useI18n} from 'vue-i18n'

import {success} from '@/message'
import {useTitle} from '@/composables/useTitle'

import {useBaseStore} from '@/stores/base'
import {useProjectStore} from '@/stores/projects'

import type { IProject } from '@/modelTypes/IProject'

definePage({
	name: 'project.settings.archive',
	meta: { showAsModal: true },
	props: route => {
		// https://github.com/posva/unplugin-vue-router/discussions/513#discussioncomment-10695660
		const castedRoute = route as RouteLocationNormalizedLoaded<'project.settings.archive'>

		return { projectId: Number(castedRoute.params.projectId) }
	},
})

const props = defineProps<{
	projectId: IProject['id'],
}>()

const {t} = useI18n({useScope: 'global'})
const projectStore = useProjectStore()
const router = useRouter()

const project = computed(() => projectStore.projects[props.projectId])
useTitle(() => t('project.archive.title', {project: project.value.title}))

async function archiveProject() {
	try {
		const newProject = await projectStore.updateProject({
			...project.value,
			isArchived: !project.value.isArchived,
		})
		useBaseStore().setCurrentProject(newProject)
		success({message: t('project.archive.success')})
	} finally {
		router.back()
	}
}
</script>
