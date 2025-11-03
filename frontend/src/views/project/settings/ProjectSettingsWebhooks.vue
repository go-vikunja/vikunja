<script lang="ts" setup>
import {ref, computed, watchEffect} from 'vue'
import {useRoute} from 'vue-router'
import {useI18n} from 'vue-i18n'
import {useTitle} from '@vueuse/core'

import ProjectService from '@/services/project'
import ProjectModel from '@/models/project'
import type {IProject} from '@/modelTypes/IProject'

import CreateEdit from '@/components/misc/CreateEdit.vue'

import {useBaseStore} from '@/stores/base'
import type {IWebhook} from '@/modelTypes/IWebhook'
import WebhookService from '@/services/webhook'
import {formatDateShort} from '@/helpers/time/formatDate'
import User from '@/components/misc/User.vue'
import WebhookModel from '@/models/webhook'
import BaseButton from '@/components/base/BaseButton.vue'
import FancyCheckbox from '@/components/input/FancyCheckbox.vue'
import {success} from '@/message'
import {isValidHttpUrl} from '@/helpers/isValidHttpUrl'

defineOptions({name: 'ProjectSettingWebhooks'})

const {t} = useI18n({useScope: 'global'})

const project = ref<IProject>()
useTitle(t('project.webhooks.title'))

const showNewForm = ref(false)

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

const webhooks = ref<IWebhook[]>()
const webhookService = new WebhookService()
const availableEvents = ref<string[]>()

async function loadWebhooks() {
	webhooks.value = await webhookService.getAll({projectId: project.value.id})
	availableEvents.value = await webhookService.getAvailableEvents()
}

const showDeleteModal = ref(false)
const webhookIdToDelete = ref<number>()

async function deleteWebhook() {
	await webhookService.delete({
		id: webhookIdToDelete.value,
		projectId: project.value.id,
	})
	showDeleteModal.value = false
	success({message: t('project.webhooks.deleteSuccess')})
	await loadWebhooks()
}

const newWebhook = ref(new WebhookModel())
const newWebhookEvents = ref({})

async function create() {

	validateTargetUrl()
	if (!webhookTargetUrlValid.value) {
		return
	}

	const selectedEvents = getSelectedEventsArray()
	newWebhook.value.events = selectedEvents

	validateSelectedEvents()
	if (!selectedEventsValid.value) {
		return
	}

	newWebhook.value.projectId = project.value.id
	const created = await webhookService.create(newWebhook.value)
	webhooks.value.push(created)
	newWebhook.value = new WebhookModel()
	showNewForm.value = false
}

const webhookTargetUrlValid = ref(true)

function validateTargetUrl() {
	webhookTargetUrlValid.value = isValidHttpUrl(newWebhook.value.targetUrl)
}

const selectedEventsValid = ref(true)

function getSelectedEventsArray() {
	return Object.entries(newWebhookEvents.value)
		.filter(([, use]) => use)
		.map(([event]) => event)
}

function validateSelectedEvents() {
	const events = getSelectedEventsArray()
	if (events.length === 0) {
		selectedEventsValid.value = false
	}
}
</script>

<template>
	<CreateEdit
		:title="$t('project.webhooks.title')"
		:has-primary-action="false"
		:wide="true"
	>
		<XButton
			v-if="!(webhooks?.length === 0 || showNewForm)"
			icon="plus"
			class="mbe-4"
			@click="showNewForm = true"
		>
			{{ $t('project.webhooks.create') }}
		</XButton>

		<div
			v-if="webhooks?.length === 0 || showNewForm"
			class="p-4"
		>
			<div class="field">
				<label
					class="label"
					for="targetUrl"
				>
					{{ $t('project.webhooks.targetUrl') }}
				</label>
				<div class="control">
					<input
						id="targetUrl"
						v-model="newWebhook.targetUrl"
						required
						class="input"
						:placeholder="$t('project.webhooks.targetUrl')"
						@focusout="validateTargetUrl"
					>
				</div>
				<p
					v-if="!webhookTargetUrlValid"
					class="help is-danger"
				>
					{{ $t('project.webhooks.targetUrlInvalid') }}
				</p>
			</div>
			<div class="field">
				<label
					class="label"
					for="secret"
				>
					{{ $t('project.webhooks.secret') }}
				</label>
				<div class="control">
					<input
						id="secret"
						v-model="newWebhook.secret"
						class="input"
					>
				</div>
				<p class="help">
					{{ $t('project.webhooks.secretHint') }}
					<BaseButton href="https://vikunja.io/docs/webhooks/">
						{{ $t('project.webhooks.secretDocs') }}
					</BaseButton>
				</p>
			</div>
			<div class="field">
				<label
					class="label"
					for="secret"
				>
					{{ $t('project.webhooks.events') }}
				</label>
				<p class="help">
					{{ $t('project.webhooks.eventsHint') }}
				</p>
				<div class="control">
					<FancyCheckbox
						v-for="event in availableEvents"
						:key="event"
						v-model="newWebhookEvents[event]"
						class="available-events-check"
						@update:modelValue="validateSelectedEvents"
					>
						{{ event }}
					</FancyCheckbox>
				</div>
				<p
					v-if="!selectedEventsValid"
					class="help is-danger"
				>
					{{ $t('project.webhooks.mustSelectEvents') }}
				</p>
			</div>
			<XButton
				icon="plus"
				@click="create"
			>
				{{ $t('project.webhooks.create') }}
			</XButton>
		</div>

		<table
			v-if="webhooks?.length > 0"
			class="table has-actions is-striped is-hoverable is-fullwidth"
		>
			<thead>
				<tr>
					<th>{{ $t('project.webhooks.targetUrl') }}</th>
					<th>{{ $t('project.webhooks.events') }}</th>
					<th>{{ $t('misc.created') }}</th>
					<th>{{ $t('misc.createdBy') }}</th>
					<th />
				</tr>
			</thead>
			<tbody>
				<tr
					v-for="w in webhooks"
					:key="w.id"
				>
					<td>{{ w.targetUrl }}</td>
					<td>{{ w.events.join(', ') }}</td>
					<td>{{ formatDateShort(w.created) }}</td>
					<td>
						<User
							:avatar-size="25"
							:user="w.createdBy"
						/>
					</td>

					<td class="actions">
						<XButton
							class="is-danger"
							icon="trash-alt"
							@click="() => {showDeleteModal = true;webhookIdToDelete = w.id}"
						/>
					</td>
				</tr>
			</tbody>
		</table>

		<Modal
			:enabled="showDeleteModal"
			@close="showDeleteModal = false"
			@submit="deleteWebhook()"
		>
			<template #header>
				<span>{{ $t('project.webhooks.delete') }}</span>
			</template>

			<template #text>
				<p>{{ $t('project.webhooks.deleteText') }}</p>
			</template>
		</Modal>
	</CreateEdit>
</template>

<style lang="scss" scoped>
.available-events-check {
	margin-inline-end: .5rem;
	inline-size: 12.5rem;
}
</style>
