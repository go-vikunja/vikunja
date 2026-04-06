<script setup lang="ts">
import {onMounted, ref} from 'vue'
import {useI18n} from 'vue-i18n'
import {useTitle} from '@/composables/useTitle'

import XButton from '@/components/input/Button.vue'
import FormField from '@/components/input/FormField.vue'
import Message from '@/components/misc/Message.vue'

import BotUserService from '@/services/botUser'
import BotTokenService from '@/services/botToken'
import UserModel from '@/models/user'
import ApiTokenModel from '@/models/apiTokenModel'
import type {IUser} from '@/modelTypes/IUser'
import type {IApiToken} from '@/modelTypes/IApiToken'

const STATUS_ACTIVE = 0
const STATUS_DISABLED = 2

const {t} = useI18n({useScope: 'global'})
useTitle(() => t('user.settings.bots.title'))

const botService = new BotUserService()
const bots = ref<IUser[]>([])
const newBotUsername = ref('')
const createError = ref<string | null>(null)

const tokensByBot = ref<Record<number, IApiToken[]>>({})
const newTokensByBot = ref<Record<number, string>>({})
const newTokenTitle = ref<Record<number, string>>({})
const newTokenExpiry = ref<Record<number, string>>({})
const showTokenForm = ref<Record<number, boolean>>({})

async function loadBots() {
	bots.value = await botService.getAll() as IUser[]
	for (const bot of bots.value) {
		await loadTokens(bot.id)
	}
}

async function loadTokens(botId: number) {
	const svc = new BotTokenService(botId)
	tokensByBot.value[botId] = await svc.getAll() as IApiToken[]
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

function toggleTokenForm(botId: number) {
	showTokenForm.value[botId] = !showTokenForm.value[botId]
	if (showTokenForm.value[botId]) {
		newTokenTitle.value[botId] = ''
		const defaultExpiry = new Date(Date.now() + 30 * 24 * 60 * 60 * 1000)
		newTokenExpiry.value[botId] = defaultExpiry.toISOString().slice(0, 16)
	}
}

async function createBotToken(bot: IUser) {
	const svc = new BotTokenService(bot.id)
	const token = new ApiTokenModel({
		title: newTokenTitle.value[bot.id] || 'default',
		permissions: {'*': '*'},
		expiresAt: new Date(newTokenExpiry.value[bot.id]),
	})
	const created = await svc.create(token) as IApiToken
	newTokensByBot.value[bot.id] = created.token
	showTokenForm.value[bot.id] = false
	await loadTokens(bot.id)
}

async function deleteBotToken(bot: IUser, token: IApiToken) {
	const svc = new BotTokenService(bot.id)
	await svc.delete(token)
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
				<XButton @click="toggleTokenForm(bot.id)">
					{{ $t('user.settings.bots.tokens.create') }}
				</XButton>
				<div
					v-if="showTokenForm[bot.id]"
					class="token-form"
				>
					<FormField :label="$t('user.settings.bots.tokens.titleLabel')">
						<input
							v-model="newTokenTitle[bot.id]"
							class="input"
							:placeholder="$t('user.settings.bots.tokens.titlePlaceholder')"
						>
					</FormField>
					<FormField :label="$t('user.settings.bots.tokens.expiryLabel')">
						<input
							v-model="newTokenExpiry[bot.id]"
							class="input"
							type="datetime-local"
						>
					</FormField>
					<XButton @click="createBotToken(bot)">
						{{ $t('user.settings.bots.tokens.createConfirm') }}
					</XButton>
				</div>
				<Message
					v-if="newTokensByBot[bot.id]"
					variant="warning"
				>
					{{ $t('user.settings.bots.tokens.showOnce') }}
					<code>{{ newTokensByBot[bot.id] }}</code>
				</Message>
				<ul>
					<li
						v-for="token in tokensByBot[bot.id] ?? []"
						:key="token.id"
					>
						{{ token.title }}
						<XButton
							variant="tertiary"
							class="is-danger is-small"
							@click="deleteBotToken(bot, token)"
						>
							{{ $t('user.settings.bots.tokens.delete') }}
						</XButton>
					</li>
				</ul>
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

.token-form {
	display: flex;
	flex-direction: column;
	gap: .5rem;
	margin-block: 1rem;
	padding: 1rem;
	border: 1px solid var(--grey-200);
	border-radius: 4px;
}

.create-form {
	display: flex;
	flex-direction: column;
	gap: .5rem;
	margin-block-end: 1rem;
}
</style>
