<script lang="ts" setup>
import {ref, computed, watchEffect} from 'vue'
import {useRoute} from 'vue-router'
import {useI18n} from 'vue-i18n'
import {useTitle} from '@vueuse/core'

import ProjectService from '@/services/project'
import ProjectModel from '@/models/project'
import type {IProject} from '@/modelTypes/IProject'
import type {IWebhook} from '@/modelTypes/IWebhook'

import CreateEdit from '@/components/misc/CreateEdit.vue'
import WebhookManager from '@/components/misc/WebhookManager.vue'

import {useBaseStore} from '@/stores/base'
import WebhookService from '@/services/webhook'
import {success} from '@/message'

defineOptions({name: 'ProjectSettingWebhooks'})

const {t} = useI18n({useScope: 'global'})

const project = ref<IProject>()
useTitle(t('project.webhooks.title'))

async function loadProject(projectId: number) {
	const projectService = new ProjectService()
	const newProject = await projectService.get(new ProjectModel({id: projectId}))
	await useBaseStore().handleSetCurrentProject({project: newProject})
	project.value = newProject
	await loadWebhooks()
}

const route = useRoute()
const projectId = computed(() => route.params.projectId !== undefined
	? parseInt(route.params.projectId as string)
	: undefined,
)

watchEffect(() => projectId.value !== undefined && loadProject(projectId.value))

const webhooks = ref<IWebhook[]>([])
const webhookService = new WebhookService()
const availableEvents = ref<string[]>([])
const loading = ref(false)

async function loadWebhooks() {
	loading.value = true
	try {
		webhooks.value = await webhookService.getAll({projectId: project.value.id})
		availableEvents.value = await webhookService.getAvailableEvents()
	} finally {
		loading.value = false
	}
}

async function handleCreate(webhook: IWebhook) {
	webhook.projectId = project.value.id
	const created = await webhookService.create(webhook)
	webhooks.value.push(created)
}

async function handleDelete(webhookId: number) {
	await webhookService.delete({
		id: webhookId,
		projectId: project.value.id,
	})
	success({message: t('project.webhooks.deleteSuccess')})
	await loadWebhooks()
}
</script>

<template>
	<CreateEdit
		:title="$t('project.webhooks.title')"
		:has-primary-action="false"
		:wide="true"
	>
		<WebhookManager
			:webhooks="webhooks"
			:available-events="availableEvents"
			:loading="loading"
			@create="handleCreate"
			@delete="handleDelete"
		/>
	</CreateEdit>
</template>
