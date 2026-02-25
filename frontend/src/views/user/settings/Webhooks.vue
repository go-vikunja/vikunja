<script lang="ts" setup>
import {ref, onMounted} from 'vue'
import {useI18n} from 'vue-i18n'

import Card from '@/components/misc/Card.vue'
import WebhookManager from '@/components/misc/WebhookManager.vue'

import {useTitle} from '@/composables/useTitle'
import {success} from '@/message'
import {UserWebhookService} from '@/services/webhook'
import type {IWebhook} from '@/modelTypes/IWebhook'

defineOptions({name: 'UserSettingsWebhooks'})

const {t} = useI18n({useScope: 'global'})
useTitle(() => `${t('user.settings.webhooks.title')} - ${t('user.settings.title')}`)

const service = new UserWebhookService()
const webhooks = ref<IWebhook[]>([])
const availableEvents = ref<string[]>([])
const loading = ref(false)

async function loadWebhooks() {
	loading.value = true
	try {
		webhooks.value = await service.getAll()
		availableEvents.value = await service.getAvailableEvents()
	} finally {
		loading.value = false
	}
}

async function handleCreate(webhook: IWebhook) {
	const created = await service.create(webhook)
	webhooks.value.push(created)
}

async function handleDelete(webhookId: number) {
	await service.delete({id: webhookId})
	success({message: t('project.webhooks.deleteSuccess')})
	await loadWebhooks()
}

onMounted(() => {
	loadWebhooks()
})
</script>

<template>
	<Card
		:title="$t('user.settings.webhooks.title')"
		:loading="loading"
	>
		<p class="mb-4">
			{{ $t('user.settings.webhooks.description') }}
		</p>

		<WebhookManager
			:webhooks="webhooks"
			:available-events="availableEvents"
			:loading="loading"
			@create="handleCreate"
			@delete="handleDelete"
		/>
	</Card>
</template>
