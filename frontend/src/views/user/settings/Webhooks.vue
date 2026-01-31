<template>
	<Card
		:title="$t('user.settings.webhooks.title')"
		class="webhook-settings"
		:loading="loading"
	>
		<p class="mb-4">
			{{ $t('user.settings.webhooks.description') }}
		</p>

		<!-- General Webhook Section -->
		<div class="webhook-section">
			<h4 class="title is-5">
				{{ $t('user.settings.webhooks.general.title') }}
			</h4>
			<p class="help mb-4">
				{{ $t('user.settings.webhooks.general.description') }}
			</p>

			<div class="field">
				<label class="two-col">
					<span>{{ $t('user.settings.webhooks.url') }}</span>
					<input
						v-model="generalWebhook.targetUrl"
						class="input"
						type="url"
						:placeholder="$t('user.settings.webhooks.urlPlaceholder')"
						@input="onGeneralUrlChange"
					>
				</label>
			</div>
			<div class="field">
				<label class="two-col">
					<span>{{ $t('user.settings.webhooks.enabled') }}</span>
					<div class="control">
						<FancyCheckbox v-model="generalWebhook.enabled" />
					</div>
				</label>
			</div>
			<div class="field">
				<XButton
					:loading="savingGeneral"
					:disabled="!generalWebhook.targetUrl"
					@click="saveGeneralWebhook"
				>
					{{ $t('misc.save') }}
				</XButton>
				<XButton
					v-if="generalWebhook.id"
					variant="tertiary"
					class="is-danger ml-2"
					:loading="deletingGeneral"
					@click="deleteGeneralWebhook"
				>
					{{ $t('misc.delete') }}
				</XButton>
			</div>
		</div>

		<!-- Specific Notification Types Section -->
		<div class="webhook-section mt-6">
			<h4 class="title is-5">
				{{ $t('user.settings.webhooks.specific.title') }}
			</h4>
			<p class="help mb-4">
				{{ $t('user.settings.webhooks.specific.description') }}
			</p>

			<div
				v-for="type in specificTypes"
				:key="type.type"
				class="webhook-type-item"
			>
				<div class="webhook-type-header">
					<strong>{{ getTypeLabel(type.type) }}</strong>
					<span class="tag">{{ type.type }}</span>
				</div>
				<p class="help">
					{{ type.description }}
				</p>

				<div class="field mt-2">
					<label class="two-col">
						<span>{{ $t('user.settings.webhooks.url') }}</span>
						<input
							v-model="webhookSettings[type.type].targetUrl"
							class="input"
							type="url"
							:placeholder="generalWebhook.targetUrl && generalWebhook.enabled
								? $t('user.settings.webhooks.usingGeneral')
								: $t('user.settings.webhooks.urlPlaceholder')"
							@input="() => onSpecificUrlChange(type.type)"
						>
					</label>
				</div>
				<div class="field">
					<label class="two-col">
						<span>{{ $t('user.settings.webhooks.enabled') }}</span>
						<div class="control">
							<FancyCheckbox v-model="webhookSettings[type.type].enabled" />
						</div>
					</label>
				</div>
				<div class="field">
					<XButton
						:loading="savingTypes[type.type]"
						:disabled="!webhookSettings[type.type].targetUrl"
						@click="() => saveSpecificWebhook(type.type)"
					>
						{{ $t('misc.save') }}
					</XButton>
					<XButton
						v-if="webhookSettings[type.type].id"
						variant="tertiary"
						class="is-danger ml-2"
						:loading="deletingTypes[type.type]"
						@click="() => deleteSpecificWebhook(type.type)"
					>
						{{ $t('misc.delete') }}
					</XButton>
				</div>
			</div>
		</div>
	</Card>
</template>

<script setup lang="ts">
import {ref, reactive, computed, onMounted} from 'vue'
import {useI18n} from 'vue-i18n'

import Card from '@/components/misc/Card.vue'
import FancyCheckbox from '@/components/input/FancyCheckbox.vue'

import {useTitle} from '@/composables/useTitle'
import {success, error} from '@/message'

import UserWebhookSettingsService from '@/services/userWebhookSettings'
import type {IUserWebhookSetting, IWebhookNotificationType} from '@/modelTypes/IUserWebhookSetting'

defineOptions({name: 'UserSettingsWebhooks'})

const {t} = useI18n({useScope: 'global'})
useTitle(() => `${t('user.settings.webhooks.title')} - ${t('user.settings.title')}`)

const service = new UserWebhookSettingsService()

const loading = ref(false)
const savingGeneral = ref(false)
const deletingGeneral = ref(false)
const savingTypes = reactive<Record<string, boolean>>({})
const deletingTypes = reactive<Record<string, boolean>>({})

const availableTypes = ref<IWebhookNotificationType[]>([])

const generalWebhook = reactive<Partial<IUserWebhookSetting>>({
	id: 0,
	notificationType: 'all',
	targetUrl: '',
	enabled: false,
})

const webhookSettings = reactive<Record<string, Partial<IUserWebhookSetting>>>({})

const specificTypes = computed(() => {
	if (!Array.isArray(availableTypes.value)) {
		return []
	}
	return availableTypes.value.filter(t => t.type !== 'all')
})

function getTypeLabel(type: string): string {
	const labels: Record<string, string> = {
		'task.reminder': t('user.settings.webhooks.types.taskReminder'),
		'task.undone.overdue': t('user.settings.webhooks.types.taskOverdue'),
	}
	return labels[type] || type
}

function onGeneralUrlChange() {
	if (generalWebhook.targetUrl && !generalWebhook.enabled) {
		generalWebhook.enabled = true
	}
}

function onSpecificUrlChange(type: string) {
	if (webhookSettings[type].targetUrl && !webhookSettings[type].enabled) {
		webhookSettings[type].enabled = true
	}
}

async function loadSettings() {
	loading.value = true
	try {
		// Load available types
		const types = await service.getAvailableTypes()
		availableTypes.value = Array.isArray(types) ? types : []

		// Initialize webhook settings for each type
		for (const type of availableTypes.value) {
			if (type.type !== 'all') {
				webhookSettings[type.type] = {
					id: 0,
					notificationType: type.type,
					targetUrl: '',
					enabled: false,
				}
			}
		}

		// Load existing settings
		const settings = await service.getAll()
		const settingsArray = Array.isArray(settings) ? settings : []
		for (const setting of settingsArray) {
			if (setting.notificationType === 'all') {
				Object.assign(generalWebhook, setting)
			} else if (webhookSettings[setting.notificationType]) {
				Object.assign(webhookSettings[setting.notificationType], setting)
			}
		}
	} catch (e) {
		error(e)
	} finally {
		loading.value = false
	}
}

async function saveGeneralWebhook() {
	savingGeneral.value = true
	try {
		const saved = await service.saveByType('all', generalWebhook.targetUrl || '', generalWebhook.enabled || false)
		Object.assign(generalWebhook, saved)
		success({message: t('user.settings.webhooks.saved')})
	} catch (e) {
		error(e)
	} finally {
		savingGeneral.value = false
	}
}

async function deleteGeneralWebhook() {
	deletingGeneral.value = true
	try {
		await service.deleteByType('all')
		generalWebhook.id = 0
		generalWebhook.targetUrl = ''
		generalWebhook.enabled = false
		success({message: t('user.settings.webhooks.deleted')})
	} catch (e) {
		error(e)
	} finally {
		deletingGeneral.value = false
	}
}

async function saveSpecificWebhook(type: string) {
	savingTypes[type] = true
	try {
		const setting = webhookSettings[type]
		const saved = await service.saveByType(type, setting.targetUrl || '', setting.enabled || false)
		Object.assign(webhookSettings[type], saved)
		success({message: t('user.settings.webhooks.saved')})
	} catch (e) {
		error(e)
	} finally {
		savingTypes[type] = false
	}
}

async function deleteSpecificWebhook(type: string) {
	deletingTypes[type] = true
	try {
		await service.deleteByType(type)
		webhookSettings[type].id = 0
		webhookSettings[type].targetUrl = ''
		webhookSettings[type].enabled = false
		success({message: t('user.settings.webhooks.deleted')})
	} catch (e) {
		error(e)
	} finally {
		deletingTypes[type] = false
	}
}

onMounted(() => {
	loadSettings()
})
</script>

<style scoped lang="scss">
.webhook-settings {
	max-width: 800px;
}

.webhook-section {
	padding: 1.5rem;
	background: var(--grey-100);
	border-radius: $radius;
	margin-bottom: 1.5rem;
}

.webhook-type-item {
	padding: 1rem;
	background: var(--white);
	border-radius: $radius;
	margin-bottom: 1rem;
	border: 1px solid var(--grey-200);

	&:last-child {
		margin-bottom: 0;
	}
}

.webhook-type-header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	margin-bottom: 0.5rem;

	.tag {
		font-family: monospace;
		font-size: 0.75rem;
		background: var(--grey-200);
		color: var(--grey-700);
		padding: 0.25rem 0.5rem;
		border-radius: $radius;
	}
}

.two-col {
	display: flex;
	align-items: center;
	gap: 1rem;

	> span:first-child {
		min-width: 120px;
	}

	> .input {
		flex: 1;
	}
}
</style>
