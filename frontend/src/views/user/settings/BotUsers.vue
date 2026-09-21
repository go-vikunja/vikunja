<script setup lang="ts">
import {computed, ref} from 'vue'
import {useI18n} from 'vue-i18n'
import {useTitle} from '@/composables/useTitle'

import XButton from '@/components/input/Button.vue'
import FormField from '@/components/input/FormField.vue'
import Message from '@/components/misc/Message.vue'
import ApiTokenForm from '@/components/token/ApiTokenForm.vue'

import {botsQuery, useCreateBotMutation, useUpdateBotMutation, useDeleteBotMutation} from '@/client/queries/bots'
import {useQueries, useQuery} from '@tanstack/vue-query'
import {botApiTokensQuery, useDeleteApiTokenMutation} from '@/client/queries/apiTokens'
import type {ApiToken, BotUser} from '@/client/generated'
import {formatDisplayDate} from '@/helpers/time/formatDate'
import {getErrorText} from '@/message'

type Bot = BotUser & Required<Pick<BotUser, 'id'>>

const STATUS_ACTIVE = 0
const STATUS_DISABLED = 2

const {t} = useI18n({useScope: 'global'})
useTitle(() => t('user.settings.bots.title'))

const {data: botData} = useQuery(botsQuery())
const createMutation = useCreateBotMutation()
const updateMutation = useUpdateBotMutation()
const deleteMutation = useDeleteBotMutation()
const deleteTokenMutation = useDeleteApiTokenMutation()
const bots = computed(() => (botData.value ?? []).filter((bot): bot is Bot => typeof bot.id === 'number' && bot.id > 0))
const newBotUsername = ref('')
const newBotName = ref('')
const createError = ref<string | null>(null)
const showCreateForm = ref(false)

const queryableBots = computed(() => bots.value.filter(bot => bot.id > 0))
const tokenQueries = useQueries({queries: computed(() => queryableBots.value.map(bot => botApiTokensQuery(bot.id)))})
const tokensByBot = computed(() => Object.fromEntries(
	queryableBots.value.map((bot, index) => [bot.id, tokenQueries.value[index]?.data ?? []]),
))
const newTokensByBot = ref<Record<number, string>>({})
const showTokenForm = ref<Record<number, boolean>>({})
const editingName = ref<Record<number, boolean>>({})
const nameDraft = ref<Record<number, string>>({})

const showDeleteModal = ref<boolean>(false)
const botToDelete = ref<Bot>()

async function createBot() {
	createError.value = null
	const username = newBotUsername.value.startsWith('bot-') ? newBotUsername.value : `bot-${newBotUsername.value}`
	const payload: Pick<BotUser, 'username' | 'name'> = {username}
	const trimmedName = newBotName.value.trim()
	if (trimmedName !== '') {
		payload.name = trimmedName
	}
	try {
		await createMutation.mutateAsync(payload)
		newBotUsername.value = ''
		newBotName.value = ''
		showCreateForm.value = false
	} catch (e: unknown) {
		createError.value = getErrorText(e)
	}
}

async function toggleBotStatus(bot: Bot) {
	const updated = {
		...bot,
		status: bot.status === STATUS_ACTIVE ? STATUS_DISABLED : STATUS_ACTIVE,
	}
	try { await updateMutation.mutateAsync({id: bot.id, body: updated}) } catch { /* Mutation reports the error. */ }
}

function startEditName(bot: Bot) {
	nameDraft.value[bot.id] = bot.name ?? ''
	editingName.value[bot.id] = true
}

function cancelEditName(bot: Bot) {
	editingName.value[bot.id] = false
	delete nameDraft.value[bot.id]
}

async function saveBotName(bot: Bot) {
	const draft = nameDraft.value[bot.id]
	const updated = {
		...bot,
		name: (draft ?? '').trim(),
	}
	try {
		await updateMutation.mutateAsync({id: bot.id, body: updated})
		if (nameDraft.value[bot.id] !== draft) {
			return
		}
		editingName.value[bot.id] = false
		delete nameDraft.value[bot.id]
	} catch { /* Mutation reports the error. */ }
}

async function deleteBot() {
	const bot = botToDelete.value
	if (!bot) {
		return
	}
	showDeleteModal.value = false
	botToDelete.value = undefined
	try {
		await deleteMutation.mutateAsync(bot.id)
		delete newTokensByBot.value[bot.id]
	} catch { /* Mutation reports the error. */ }
}

function onTokenCreated(bot: Bot, token: ApiToken) {
	newTokensByBot.value[bot.id] = token.token ?? ''
	showTokenForm.value[bot.id] = false
}

async function deleteToken(token: ApiToken) {
	if (!token.id) return
	try { await deleteTokenMutation.mutateAsync(token.id) } catch { /* Mutation reports the error. */ }
}

</script>

<template>
	<div class="content">
		<h2>{{ $t('user.settings.bots.title') }}</h2>
		<p>{{ $t('user.settings.bots.description') }}</p>

		<div
			v-if="bots.length === 0 || showCreateForm"
			class="create-form"
		>
			<FormField
				:label="$t('user.auth.username')"
				:error="createError"
			>
				<input
					v-model="newBotUsername"
					class="input"
					placeholder="bot-myassistant"
				>
			</FormField>
			<FormField :label="$t('admin.users.nameLabel')">
				<input
					v-model="newBotName"
					class="input"
					:placeholder="$t('user.settings.bots.namePlaceholder')"
				>
			</FormField>
			<XButton
				:loading="createMutation.isPending.value"
				@click="createBot"
			>
				{{ $t('user.settings.bots.create') }}
			</XButton>
		</div>
		<XButton
			v-else
			icon="plus"
			class="mbe-4"
			@click="showCreateForm = true"
		>
			{{ $t('user.settings.bots.create') }}
		</XButton>

		<div
			v-for="bot in bots"
			:key="bot.id"
			class="bot-card"
		>
			<div class="bot-header">
				<strong>{{ bot.username }}</strong>
				<template v-if="editingName[bot.id]">
					<span class="bot-name-edit">—</span>
					<input
						v-model="nameDraft[bot.id]"
						v-focus
						class="input bot-name-input"
						:placeholder="$t('user.settings.bots.namePlaceholder')"
						@keyup.enter="saveBotName(bot)"
						@keyup.esc="cancelEditName(bot)"
					>
					<XButton
						variant="secondary"
						@click="saveBotName(bot)"
					>
						{{ $t('misc.save') }}
					</XButton>
					<XButton
						variant="tertiary"
						@click="cancelEditName(bot)"
					>
						{{ $t('misc.cancel') }}
					</XButton>
				</template>
				<template v-else>
					<span v-if="bot.name">— {{ bot.name }}</span>
					<span
						v-else
						class="no-name"
					>{{ $t('project.share.links.noName') }}</span>
					<XButton
						variant="tertiary"
						icon="pencil-alt"
						@click="startEditName(bot)"
					>
						{{ $t('menu.edit') }}
					</XButton>
				</template>
				<span class="status">{{ bot.status === STATUS_ACTIVE ? $t('admin.users.statusActive') : $t('admin.users.statusDisabled') }}</span>
			</div>
			<div class="bot-actions">
				<XButton
					variant="secondary"
					@click="toggleBotStatus(bot)"
				>
					{{ bot.status === STATUS_ACTIVE ? $t('misc.disable') : $t('user.settings.bots.enable') }}
				</XButton>
				<XButton
					variant="tertiary"
					class="is-danger"
					@click="() => {botToDelete = bot; showDeleteModal = true}"
				>
					{{ $t('misc.delete') }}
				</XButton>
			</div>

			<div class="tokens">
				<h4>{{ $t('user.settings.apiTokens.title') }}</h4>
				<Message
					v-if="newTokensByBot[bot.id]"
					variant="warning"
				>
					{{ $t('user.settings.apiTokens.tokenCreatedNotSeeAgain') }}
					<code>{{ newTokensByBot[bot.id] }}</code>
				</Message>
				<div
					v-if="(tokensByBot[bot.id] ?? []).length > 0"
					class="has-horizontal-overflow"
				>
					<table class="table">
						<thead>
							<tr>
								<th>{{ $t('user.settings.apiTokens.attributes.title') }}</th>
								<th>{{ $t('user.settings.apiTokens.attributes.expiresAt') }}</th>
								<th>{{ $t('misc.created') }}</th>
								<th class="has-text-end">
									{{ $t('misc.actions') }}
								</th>
							</tr>
						</thead>
						<tbody>
							<tr
								v-for="token in tokensByBot[bot.id] ?? []"
								:key="token.id"
							>
								<td>{{ token.title }}</td>
								<td>{{ formatDisplayDate(token.expires_at) }}</td>
								<td>{{ formatDisplayDate(token.created) }}</td>
								<td class="has-text-end">
									<XButton
										variant="secondary"
										@click="deleteToken(token)"
									>
										{{ $t('misc.delete') }}
									</XButton>
								</td>
							</tr>
						</tbody>
					</table>
				</div>
				<template v-if="bot.id > 0">
					<ApiTokenForm
						v-if="showTokenForm[bot.id]"
						:owner-id="bot.id"
						@created="(token: ApiToken) => onTokenCreated(bot, token)"
						@cancel="showTokenForm[bot.id] = false"
					/>
					<XButton
						v-else
						icon="plus"
						class="mbe-4"
						@click="showTokenForm[bot.id] = true"
					>
						{{ $t('user.settings.apiTokens.createToken') }}
					</XButton>
				</template>
			</div>
		</div>

		<Modal
			:enabled="showDeleteModal"
			@close="showDeleteModal = false"
			@submit="deleteBot()"
		>
			<template #header>
				{{ $t('user.settings.bots.delete.header') }}
			</template>

			<template #text>
				<p v-if="botToDelete">
					{{ $t('user.settings.bots.delete.text1', {username: botToDelete.username}) }}<br>
					{{ $t('user.settings.bots.delete.text2') }}
				</p>
			</template>
		</Modal>
	</div>
</template>

<style lang="scss" scoped>
.bot-card {
	padding: 1rem;
	margin-block-start: 1rem;
	border: 1px solid var(--grey-200);
	border-radius: 4px;
}

.bot-header {
	display: flex;
	gap: .5rem;
	align-items: center;
	margin-block-end: .5rem;
}

.bot-name-input {
	max-inline-size: 16rem;
}

.no-name {
	font-style: italic;
	color: var(--grey-500);
}

.status {
	margin-inline-start: auto;
	font-size: .85rem;
	color: var(--grey-600);
}

.bot-actions {
	display: flex;
	gap: .5rem;
	margin-block-end: 1rem;
}

.tokens {
	margin-block-start: 1rem;
	padding-block-start: 1rem;
	border-block-start: 1px solid var(--grey-200);
}

.create-form {
	display: flex;
	flex-direction: column;
	gap: .5rem;
	margin-block-end: 1rem;
}
</style>
