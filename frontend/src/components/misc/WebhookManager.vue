<script lang="ts" setup>
import {computed, ref, watch} from 'vue'

import type {WebhookWritable} from '@/client/generated'
import {useQuery} from '@tanstack/vue-query'
import {webhooksQuery, webhookEventsQuery, useCreateWebhookMutation, useDeleteWebhookMutation, type WebhookScope} from '@/client/queries/webhooks'
import BaseButton from '@/components/base/BaseButton.vue'
import FancyCheckbox from '@/components/input/FancyCheckbox.vue'
import FormField from '@/components/input/FormField.vue'
import FormInput from '@/components/input/FormInput.vue'
import Expandable from '@/components/base/Expandable.vue'
import User from '@/components/misc/User.vue'
import {formatDateShort} from '@/helpers/time/formatDate'
import {isValidHttpUrl} from '@/helpers/isValidHttpUrl'

const props = defineProps<{scope: WebhookScope}>()
const {data: webhookData, isFetching, isPending} = useQuery(computed(() => webhooksQuery(props.scope)))
const {data: eventData} = useQuery(computed(() => webhookEventsQuery(props.scope.kind)))
const webhooks = computed(() => webhookData.value ?? [])
const availableEvents = computed(() => eventData.value ?? [])
const createMutation = useCreateWebhookMutation()
const deleteMutation = useDeleteWebhookMutation()
const loading = computed(() => isFetching.value || createMutation.isPending.value || deleteMutation.isPending.value)
function emptyDraft(): WebhookWritable & {target_url: string} {
	return {target_url: '', secret: '', basic_auth_user: '', basic_auth_password: '', events: []}
}

defineOptions({name: 'WebhookManager'})

const showNewForm = ref(false)
const showBasicAuth = ref(false)
const newWebhook = ref(emptyDraft())
const newWebhookEvents = ref<Record<string, boolean>>({})

function initEvents(events: string[]) {
	newWebhookEvents.value = Object.fromEntries(
		events.map(event => [event, false]),
	)
}

watch(availableEvents, (events) => {
	if (events) initEvents(events)
}, {immediate: true})

const webhookTargetUrlValid = ref(true)
const selectedEventsValid = ref(true)
const showDeleteModal = ref(false)
const webhookIdToDelete = ref<number>()

function validateTargetUrl() {
	webhookTargetUrlValid.value = isValidHttpUrl(newWebhook.value.target_url)
}

function getSelectedEventsArray() {
	return Object.entries(newWebhookEvents.value)
		.filter(([, use]) => use)
		.map(([event]) => event)
}

function validateSelectedEvents() {
	const events = getSelectedEventsArray()
	selectedEventsValid.value = events.length > 0
}

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

	try {
		await createMutation.mutateAsync({scope: props.scope, body: newWebhook.value})
	} catch { return }
	newWebhook.value = emptyDraft()
	initEvents(availableEvents.value)
	showNewForm.value = false
}

function confirmDelete(webhookId: number | undefined) {
	webhookIdToDelete.value = webhookId
	showDeleteModal.value = true
}

function doDelete() {
	if (webhookIdToDelete.value) {
		deleteMutation.mutate({scope: props.scope, id: webhookIdToDelete.value})
	}
	showDeleteModal.value = false
}
</script>

<template>
	<div
		class="loader-container"
		:class="{'is-loading': isPending}"
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
			v-if="!isPending && (webhooks?.length === 0 || showNewForm)"
			class="p-4"
		>
			<FormField
				id="targetUrl"
				v-model="newWebhook.target_url"
				:label="$t('project.webhooks.targetUrl')"
				required
				:placeholder="$t('project.webhooks.targetUrl')"
				:error="webhookTargetUrlValid ? null : $t('project.webhooks.targetUrlInvalid')"
				@focusout="validateTargetUrl"
			/>
			<FormField :label="$t('project.webhooks.secret')">
				<template #default="{id}">
					<FormInput
						:id="id"
						v-model="newWebhook.secret"
					/>
				</template>
			</FormField>
			<p class="help">
				{{ $t('project.webhooks.secretHint') }}
				<BaseButton href="https://vikunja.io/docs/webhooks/">
					{{ $t('project.webhooks.secretDocs') }}
				</BaseButton>
			</p>
			<BaseButton
				class="mbe-2 has-text-primary"
				@click="showBasicAuth = !showBasicAuth"
			>
				{{ $t('project.webhooks.basicauthlink') }}
			</BaseButton>
			<Expandable
				:open="showBasicAuth"
				class="content"
			>
				<FormField :label="$t('project.webhooks.basicauthuser')">
					<template #default="{id}">
						<FormInput
							:id="id"
							v-model="newWebhook.basic_auth_user"
						/>
					</template>
				</FormField>
				<FormField :label="$t('project.webhooks.basicauthpassword')">
					<template #default="{id}">
						<FormInput
							:id="id"
							v-model="newWebhook.basic_auth_password"
						/>
					</template>
				</FormField>
			</Expandable>
			<div class="field">
				<label
					class="label"
					for="events"
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
				:loading="loading"
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
					<td class="webhook-target-url">
						{{ w.target_url }}
					</td>
					<td>{{ (w.events ?? []).join(', ') }}</td>
					<td>{{ formatDateShort(w.created) }}</td>
					<td>
						<User
							:avatar-size="25"
							:user="w.created_by ?? {}"
						/>
					</td>

					<td class="actions">
						<XButton
							danger
							icon="trash-alt"
							:aria-label="$t('project.webhooks.delete')"
							@click="() => confirmDelete(w.id)"
						/>
					</td>
				</tr>
			</tbody>
		</table>

		<Modal
			:enabled="showDeleteModal"
			@close="showDeleteModal = false"
			@submit="doDelete()"
		>
			<template #header>
				<span>{{ $t('project.webhooks.delete') }}</span>
			</template>

			<template #text>
				<p>{{ $t('project.webhooks.deleteText') }}</p>
			</template>
		</Modal>
	</div>
</template>

<style lang="scss" scoped>
.available-events-check {
	margin-inline-end: .5rem;
	inline-size: 12.5rem;
}

// Webhook URLs have no break opportunities, so without this the cell's min-content
// width forces the whole table wider than its container.
.webhook-target-url {
	overflow-wrap: anywhere;
}
</style>
