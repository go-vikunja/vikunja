<template>
	<create-edit
		:title="$t('project.share.header')"
		:has-primary-action="false"
	>
		<template v-if="project">
			<userTeam
				:id="project.id"
				:userIsAdmin="userIsAdmin"
				shareType="user"
				type="project"
			/>
			<userTeam
				:id="project.id"
				:userIsAdmin="userIsAdmin"
				shareType="team"
				type="project"
			/>
		</template>

		<link-sharing :project-id="projectId" v-if="linkSharingEnabled" class="mt-4"/>
	</create-edit>
</template>

<script lang="ts">
export default {name: 'project-setting-share'}
</script>

<script lang="ts" setup>
import {ref, computed, watchEffect} from 'vue'
import {useRoute} from 'vue-router'
import {useI18n} from 'vue-i18n'
import {useTitle} from '@vueuse/core'

import ProjectService from '@/services/project'
import ProjectModel from '@/models/project'
import type {IProject} from '@/modelTypes/IProject'
import {RIGHTS} from '@/constants/rights'

import CreateEdit from '@/components/misc/create-edit.vue'
import LinkSharing from '@/components/sharing/linkSharing.vue'
import userTeam from '@/components/sharing/userTeam.vue'

import {useBaseStore} from '@/stores/base'
import {useConfigStore} from '@/stores/config'

const {t} = useI18n({useScope: 'global'})

const project = ref<IProject>()
const title = computed(() => project.value?.title
	? t('project.share.title', {project: project.value.title})
	: '',
)
useTitle(title)

const configStore = useConfigStore()

const linkSharingEnabled = computed(() => configStore.linkSharingEnabled)
const userIsAdmin = computed(() => project?.value?.maxRight === RIGHTS.ADMIN)

async function loadProject(projectId: number) {
	const projectService = new ProjectService()
	const newProject = await projectService.get(new ProjectModel({id: projectId}))
	await useBaseStore().handleSetCurrentProject({project: newProject})
	project.value = newProject
}

const route = useRoute()
const projectId = computed(() => route.params.projectId !== undefined
	? parseInt(route.params.projectId as string)
	: undefined,
)
watchEffect(() => projectId.value !== undefined && loadProject(projectId.value))
</script>
