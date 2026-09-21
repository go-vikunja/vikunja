<template>
	<CreateEdit
		:title="$t('project.share.header')"
		:has-primary-action="false"
	>
		<template v-if="project">
			<UserSharing
				:project-id="project.id"
				:user-is-admin="userIsAdmin"
			/>
			<TeamSharing
				:project-id="project.id"
				:user-is-admin="userIsAdmin"
			/>
		</template>

		<LinkSharing
			v-if="link_sharing_enabled && userIsAdmin"
			:project-id="projectId"
			class="mbs-4"
		/>
	</CreateEdit>
</template>


<script lang="ts" setup>
import {computed, watchEffect} from 'vue'
import {useQuery} from '@tanstack/vue-query'
import {useRoute} from 'vue-router'
import {useI18n} from 'vue-i18n'
import {useTitle} from '@vueuse/core'

import {PERMISSIONS} from '@/constants/permissions'

import CreateEdit from '@/components/misc/CreateEdit.vue'
import LinkSharing from '@/components/sharing/LinkSharing.vue'
import TeamSharing from '@/components/sharing/TeamSharing.vue'
import UserSharing from '@/components/sharing/UserSharing.vue'

import {useBaseStore} from '@/stores/base'
import {useConfigStore} from '@/stores/config'
import {projectQuery} from '@/client/queries/projects'

defineOptions({name: 'ProjectSettingShare'})

const {t} = useI18n({useScope: 'global'})

const route = useRoute()
const projectId = computed(() => Number(route.params.projectId))
const {data: project} = useQuery(computed(() => ({...projectQuery(projectId.value), enabled: projectId.value > 0})))
const title = computed(() => project.value?.title
	? t('project.share.title', {project: project.value.title})
	: '',
)
useTitle(title)

const configStore = useConfigStore()

const link_sharing_enabled = computed(() => configStore.link_sharing_enabled)
const userIsAdmin = computed(() => project.value?.max_permission === PERMISSIONS.ADMIN)

const baseStore = useBaseStore()
watchEffect(() => {
	if (project.value) baseStore.setCurrentProject(project.value)
})
</script>
