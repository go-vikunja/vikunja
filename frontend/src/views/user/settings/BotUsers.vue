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
const createError = ref<string | null>(null)

const tokensByBot = ref<Record<number, IApiToken[]>>({})
const newTokensByBot = ref<Record<number, string>>({})
const showTokenForm = ref<Record<number, boolean>>({})

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
	if (!newBotUsername.value.startsWith('bot-')) {
		createError.value = t('user.settings.bots.prefixRequired')
		return
	}
	try {
		const created = await botService.create(new UserModel({username: newBotUsername.value}))
		bots.value.push(created as IUser)
		newBotUsername.value = ''
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

async function deleteBot(bot: IUser) {
	if (!window.confirm(t('user.settings.bots.deleteConfirm', {username: bot.username}))) {
		return
	}
	await botService.delete(bot)
	bots.value = bots.value.filter(b => b.id !== bot.id)
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

		<div class="create-form">
			<FormField
				:label="$t('user.settings.bots.usernameLabel')"
				:error="createError"
			>
				<input
					v-model="newBotUsername"
					class="input"
					:placeholder="$t('user.settings.bots.usernamePlaceholder')"
				>
			</FormField>
			<XButton @click="createBot">
				{{ $t('user.settings.bots.create') }}
			</XButton>
		</div>

		<div
			v-for="bot in bots"
			:key="bot.id"
			class="bot-card"
		>
			<div class="bot-header">
				<strong>{{ bot.username }}</strong>
				<span v-if="bot.name">— {{ bot.name }}</span>
				<span class="status">{{ bot.status === STATUS_ACTIVE ? $t('user.settings.bots.statusActive') : $t('user.settings.bots.statusDisabled') }}</span>
			</div>
			<div class="bot-actions">
				<XButton
					variant="secondary"
					@click="toggleBotStatus(bot)"
				>
					{{ bot.status === STATUS_ACTIVE ? $t('user.settings.bots.disable') : $t('user.settings.bots.enable') }}
				</XButton>
				<XButton
					variant="tertiary"
					class="is-danger"
					@click="deleteBot(bot)"
				>
					{{ $t('user.settings.bots.delete') }}
				</XButton>
			</div>

			<div class="tokens">
				<h4>{{ $t('user.settings.bots.tokens.title') }}</h4>
				<Message
					v-if="newTokensByBot[bot.id]"
					variant="warning"
				>
					{{ $t('user.settings.bots.tokens.showOnce') }}
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
										{{ $t('user.settings.bots.tokens.delete') }}
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
				/>
				<XButton
					v-else
					icon="plus"
					class="mbe-4"
					@click="showTokenForm[bot.id] = true"
				>
					{{ $t('user.settings.bots.tokens.create') }}
				</XButton>
			</div>
		</div>
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
