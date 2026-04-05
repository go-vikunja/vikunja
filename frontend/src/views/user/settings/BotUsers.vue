<script setup lang="ts">
import {onMounted, ref} from 'vue'
import {useI18n} from 'vue-i18n'
import {useTitle} from '@/composables/useTitle'

import XButton from '@/components/input/Button.vue'
import FormField from '@/components/input/FormField.vue'
import Message from '@/components/misc/Message.vue'
import ApiTokenForm from '@/components/token/ApiTokenForm.vue'

import BotUserService from '@/services/botUser'
import ApiTokenService from '@/services/apiToken'
import UserModel from '@/models/user'
import type {IUser} from '@/modelTypes/IUser'
import type {IApiToken} from '@/modelTypes/IApiToken'
import {formatDisplayDate} from '@/helpers/time/formatDate'

const STATUS_ACTIVE = 0
const STATUS_DISABLED = 2

const {t} = useI18n({useScope: 'global'})
useTitle(() => t('user.settings.bots.title'))

const botService = new BotUserService()
const tokenService = new ApiTokenService()
const bots = ref<IUser[]>([])
const newBotUsername = ref('')
const newBotName = ref('')
const createError = ref<string | null>(null)
const showCreateForm = ref(false)

const tokensByBot = ref<Record<number, IApiToken[]>>({})
const newTokensByBot = ref<Record<number, string>>({})
const showTokenForm = ref<Record<number, boolean>>({})
const editingName = ref<Record<number, boolean>>({})
const nameDraft = ref<Record<number, string>>({})

const showDeleteModal = ref<boolean>(false)
const botToDelete = ref<IUser>()

async function loadBots() {
	bots.value = await botService.getAll() as IUser[]
	for (const bot of bots.value) {
		await loadTokens(bot.id)
	}
}

async function loadTokens(botId: number) {
	tokensByBot.value[botId] = await tokenService.getAll({}, {owner_id: botId}) as IApiToken[]
}

async function createBot() {
	createError.value = null
	const username = newBotUsername.value.startsWith('bot-') ? newBotUsername.value : `bot-${newBotUsername.value}`
	const payload: Partial<IUser> = {username}
	const trimmedName = newBotName.value.trim()
	if (trimmedName !== '') {
		payload.name = trimmedName
	}
	try {
		const created = await botService.create(new UserModel(payload))
		bots.value.push(created as IUser)
		newBotUsername.value = ''
		newBotName.value = ''
		showCreateForm.value = false
	} catch (e: unknown) {
		const err = e as {response?: {data?: {message?: string}}}
		createError.value = err?.response?.data?.message ?? String(e)
	}
}

async function toggleBotStatus(bot: IUser) {
	const updated = new UserModel({
		...bot,
		status: bot.status === STATUS_ACTIVE ? STATUS_DISABLED : STATUS_ACTIVE,
	})
	const result = await botService.update(updated) as IUser
	const idx = bots.value.findIndex(b => b.id === bot.id)
	if (idx >= 0) {
		bots.value[idx] = result
	}
}

function startEditName(bot: IUser) {
	nameDraft.value[bot.id] = bot.name ?? ''
	editingName.value[bot.id] = true
}

function cancelEditName(bot: IUser) {
	editingName.value[bot.id] = false
	delete nameDraft.value[bot.id]
}

async function saveBotName(bot: IUser) {
	const updated = new UserModel({
		...bot,
		name: (nameDraft.value[bot.id] ?? '').trim(),
	})
	const result = await botService.update(updated) as IUser
	const idx = bots.value.findIndex(b => b.id === bot.id)
	if (idx >= 0) {
		bots.value[idx] = result
	}
	editingName.value[bot.id] = false
	delete nameDraft.value[bot.id]
}

async function deleteBot() {
	const bot = botToDelete.value
	if (!bot) {
		return
	}
	await botService.delete(bot)
	bots.value = bots.value.filter(b => b.id !== bot.id)
	showDeleteModal.value = false
	botToDelete.value = undefined
}

function onTokenCreated(bot: IUser, token: IApiToken) {
	newTokensByBot.value[bot.id] = token.token
	showTokenForm.value[bot.id] = false
	loadTokens(bot.id)
}

async function deleteToken(bot: IUser, token: IApiToken) {
	await tokenService.delete(token)
	await loadTokens(bot.id)
}

onMounted(loadBots)
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
			<XButton @click="createBot">
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
								<td>{{ formatDisplayDate(token.expiresAt) }}</td>
								<td>{{ formatDisplayDate(token.created) }}</td>
								<td class="has-text-end">
									<XButton
										variant="secondary"
										@click="deleteToken(bot, token)"
									>
										{{ $t('misc.delete') }}
									</XButton>
								</td>
							</tr>
						</tbody>
					</table>
				</div>
				<ApiTokenForm
					v-if="showTokenForm[bot.id]"
					:owner-id="bot.id"
					@created="(token: IApiToken) => onTokenCreated(bot, token)"
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
				<p>
					{{ $t('user.settings.bots.delete.text1', {username: botToDelete?.username}) }}<br>
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
