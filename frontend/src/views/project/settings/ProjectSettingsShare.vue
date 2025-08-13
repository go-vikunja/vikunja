<template>
	<CreateEdit
		:title="$t('project.share.header')"
		:has-primary-action="false"
	>
		<template v-if="project">
			<userTeam
				:id="project.id"
				:user-is-admin="userIsAdmin"
				share-type="user"
			/>
			<userTeam
				:id="project.id"
				:user-is-admin="userIsAdmin"
				share-type="team"
			/>
		</template>

		<LinkSharing
			v-if="linkSharingEnabled"
			:project-id="projectId"
			class="mbs-4"
		/>
	</CreateEdit>
</template>


<script lang="ts" setup>
import {ref, computed, watchEffect} from 'vue'
import {useRoute} from 'vue-router'
import {useI18n} from 'vue-i18n'
import {useTitle} from '@vueuse/core'

import ProjectService from '@/services/project'
import ProjectModel from '@/models/project'
import type {IProject} from '@/modelTypes/IProject'
import {PERMISSIONS} from '@/constants/permissions'

import CreateEdit from '@/components/misc/CreateEdit.vue'
import LinkSharing from '@/components/sharing/LinkSharing.vue'
import userTeam from '@/components/sharing/UserTeam.vue'

import {useBaseStore} from '@/stores/base'
import {useConfigStore} from '@/stores/config'

defineOptions({name: 'ProjectSettingShare'})

const {t} = useI18n({useScope: 'global'})

const project = ref<IProject>()
const title = computed(() => project.value?.title
	? t('project.share.title', {project: project.value.title})
	: '',
)
useTitle(title)

const configStore = useConfigStore()

const linkSharingEnabled = computed(() => configStore.linkSharingEnabled)
const userIsAdmin = computed(() => project?.value?.maxPermission === PERMISSIONS.ADMIN)

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
