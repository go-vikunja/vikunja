<script setup lang="ts">
import {onMounted, ref} from 'vue'
import {useI18n} from 'vue-i18n'
import {useTitle} from '@/composables/useTitle'

import XButton from '@/components/input/Button.vue'
import FormField from '@/components/input/FormField.vue'
import Message from '@/components/misc/Message.vue'

import BotUserService from '@/services/botUser'
import BotTokenService from '@/services/botToken'
import BotUserModel from '@/models/botUser'
import ApiTokenModel from '@/models/apiTokenModel'
import type {IBotUser} from '@/modelTypes/IBotUser'
import type {IApiToken} from '@/modelTypes/IApiToken'

const {t} = useI18n({useScope: 'global'})
useTitle(() => t('user.settings.bots.title'))

const botService = new BotUserService()
const bots = ref<IBotUser[]>([])
const newBotUsername = ref('')
const createError = ref('')

const tokensByBot = ref<Record<number, IApiToken[]>>({})
const newTokensByBot = ref<Record<number, string>>({})

async function loadBots() {
	bots.value = await botService.getAll() as IBotUser[]
	for (const bot of bots.value) {
		await loadTokens(bot.id)
	}
}

async function loadTokens(botId: number) {
	const svc = new BotTokenService(botId)
	tokensByBot.value[botId] = await svc.getAll() as IApiToken[]
}

async function createBot() {
	createError.value = ''
	if (!newBotUsername.value.startsWith('bot-')) {
		createError.value = t('user.settings.bots.prefixRequired')
		return
	}
	try {
		const created = await botService.create(new BotUserModel({username: newBotUsername.value}))
		bots.value.push(created as IBotUser)
		newBotUsername.value = ''
	} catch (e: unknown) {
		const err = e as {response?: {data?: {message?: string}}}
		createError.value = err?.response?.data?.message ?? String(e)
	}
}

async function toggleBotStatus(bot: IBotUser) {
	const updated = new BotUserModel({
		...bot,
		status: bot.status === 0 ? 2 : 0,
	})
	const result = await botService.update(updated) as IBotUser
	const idx = bots.value.findIndex(b => b.id === bot.id)
	if (idx >= 0) {
		bots.value[idx] = result
	}
}

async function deleteBot(bot: IBotUser) {
	if (!window.confirm(t('user.settings.bots.deleteConfirm', {username: bot.username}))) {
		return
	}
	await botService.delete(bot)
	bots.value = bots.value.filter(b => b.id !== bot.id)
}

async function createBotToken(bot: IBotUser) {
	const svc = new BotTokenService(bot.id)
	const token = new ApiTokenModel({
		title: 'default',
		permissions: {'tasks': ['read_all']},
		expiresAt: new Date(Date.now() + 365 * 24 * 60 * 60 * 1000),
	})
	const created = await svc.create(token) as IApiToken
	newTokensByBot.value[bot.id] = created.token
	await loadTokens(bot.id)
}

async function deleteBotToken(bot: IBotUser, token: IApiToken) {
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
			<FormField :label="$t('user.settings.bots.usernameLabel')">
				<input
					v-model="newBotUsername"
					class="input"
					:placeholder="$t('user.settings.bots.usernamePlaceholder')"
				>
			</FormField>
			<XButton @click="createBot">
				{{ $t('user.settings.bots.create') }}
			</XButton>
			<Message
				v-if="createError"
				variant="danger"
			>
				{{ createError }}
			</Message>
		</div>

		<div
			v-for="bot in bots"
			:key="bot.id"
			class="bot-card"
		>
			<div class="bot-header">
				<strong>{{ bot.username }}</strong>
				<span v-if="bot.name">— {{ bot.name }}</span>
				<span class="status">{{ bot.status === 0 ? $t('user.settings.bots.statusActive') : $t('user.settings.bots.statusDisabled') }}</span>
			</div>
			<div class="bot-actions">
				<XButton
					variant="secondary"
					@click="toggleBotStatus(bot)"
				>
					{{ bot.status === 0 ? $t('user.settings.bots.disable') : $t('user.settings.bots.enable') }}
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
				<XButton @click="createBotToken(bot)">
					{{ $t('user.settings.bots.tokens.create') }}
				</XButton>
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

.create-form {
	display: flex;
	flex-direction: column;
	gap: .5rem;
	margin-block-end: 1rem;
}
</style>
