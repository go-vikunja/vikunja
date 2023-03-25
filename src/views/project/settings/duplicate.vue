<template>
	<create-edit
		:title="$t('project.duplicate.title')"
		primary-icon="paste"
		:primary-label="$t('project.duplicate.label')"
		@primary="duplicateProject"
		:loading="projectDuplicateService.loading"
	>
		<p>{{ $t('project.duplicate.text') }}</p>
	</create-edit>
</template>

<script setup lang="ts">
import {ref, shallowReactive} from 'vue'
import {useRoute, useRouter} from 'vue-router'
import {useI18n} from 'vue-i18n'

import ProjectDuplicateService from '@/services/projectDuplicateService'
import CreateEdit from '@/components/misc/create-edit.vue'
import Multiselect from '@/components/input/multiselect.vue'

import ProjectDuplicateModel from '@/models/projectDuplicateModel'

import {success} from '@/message'
import {useTitle} from '@/composables/useTitle'
import {useProjectStore} from '@/stores/projects'

const {t} = useI18n({useScope: 'global'})
useTitle(() => t('project.duplicate.title'))

const route = useRoute()
const router = useRouter()
const projectStore = useProjectStore()

const projectDuplicateService = shallowReactive(new ProjectDuplicateService())

async function duplicateProject() {
	const projectDuplicate = new ProjectDuplicateModel({
		// FIXME: should be parameter
		projectId: route.params.projectId,
	})

	const duplicate = await projectDuplicateService.create(projectDuplicate)

	projectStore.setProject(duplicate.project)
	success({message: t('project.duplicate.success')})
	router.push({name: 'project.index', params: {projectId: duplicate.project.id}})
}
</script>
