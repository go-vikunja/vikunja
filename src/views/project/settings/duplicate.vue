<template>
	<create-edit
		:title="$t('project.duplicate.title')"
		primary-icon="paste"
		:primary-label="$t('project.duplicate.label')"
		@primary="duplicateProject"
		:loading="projectDuplicateService.loading"
	>
		<p>{{ $t('project.duplicate.text') }}</p>

		<Multiselect
			:placeholder="$t('namespace.search')"
			@search="findNamespaces"
			:search-results="namespaces"
			@select="selectNamespace"
			label="title"
			:search-delay="10"
		/>
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
import type {INamespace} from '@/modelTypes/INamespace'

import {success} from '@/message'
import {useTitle} from '@/composables/useTitle'
import {useNamespaceSearch} from '@/composables/useNamespaceSearch'
import {useProjectStore} from '@/stores/projects'
import {useNamespaceStore} from '@/stores/namespaces'

const {t} = useI18n({useScope: 'global'})
useTitle(() => t('project.duplicate.title'))

const {
	namespaces,
	findNamespaces,
} = useNamespaceSearch()

const selectedNamespace = ref<INamespace>()

function selectNamespace(namespace: INamespace) {
	selectedNamespace.value = namespace
}

const route = useRoute()
const router = useRouter()
const projectStore = useProjectStore()
const namespaceStore = useNamespaceStore()

const projectDuplicateService = shallowReactive(new ProjectDuplicateService())

async function duplicateProject() {
	const projectDuplicate = new ProjectDuplicateModel({
		// FIXME: should be parameter
		projectId: route.params.projectId,
		namespaceId: selectedNamespace.value?.id,
	})

	const duplicate = await projectDuplicateService.create(projectDuplicate)

	namespaceStore.addProjectToNamespace(duplicate.project)
	projectStore.setProject(duplicate.project)
	success({message: t('project.duplicate.success')})
	router.push({name: 'project.index', params: {projectId: duplicate.project.id}})
}
</script>
