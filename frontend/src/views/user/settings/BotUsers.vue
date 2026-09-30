<script setup lang="ts">
import {computed, ref, watch} from 'vue'
import {useI18n} from 'vue-i18n'
import {useRouteQuery} from '@vueuse/router'
import {useTitle} from '@/composables/useTitle'
import {useClampedPage} from '@/composables/useClampedPage'

import XButton from '@/components/input/Button.vue'
import FormField from '@/components/input/FormField.vue'
import Message from '@/components/misc/Message.vue'
import Pagination from '@/components/misc/Pagination.vue'
import PaginationEmit from '@/components/misc/PaginationEmit.vue'
import ApiTokenForm from '@/components/token/ApiTokenForm.vue'

import {botsQuery, useCreateBotMutation, useUpdateBotMutation, useDeleteBotMutation} from '@/client/queries/bots'
import {clampPage, normalizePageNumber} from '@/client/queries/pagination'
import {useQueries, useQuery} from '@tanstack/vue-query'
import {botApiTokensQuery, useDeleteApiTokenMutation} from '@/client/queries/apiTokens'
import type {ApiToken, BotUser} from '@/client/generated'
import {formatDisplayDate} from '@/helpers/time/formatDate'
import {getErrorText} from '@/message'

type Bot = BotUser & Required<Pick<BotUser, 'id' | 'status'>>

const STATUS_ACTIVE = 0
const STATUS_DISABLED = 2

const {t} = useI18n({useScope: 'global'})
useTitle(() => t('user.settings.bots.title'))

const page = useRouteQuery('page', '1', {transform: normalizePageNumber})
const {data: botPage, isPlaceholderData} = useQuery(computed(() => botsQuery(page.value)))
const createMutation = useCreateBotMutation()
const updateMutation = useUpdateBotMutation()
const deleteMutation = useDeleteBotMutation()
const deleteTokenMutation = useDeleteApiTokenMutation()
const bots = computed(() => (botPage.value?.items ?? []).filter((bot): bot is Bot => typeof bot.id === 'number' && bot.id > 0 && typeof bot.status === 'number'))
const totalPages = computed(() => botPage.value?.total_pages ?? 0)
const hasBots = computed(() => (botPage.value?.total ?? 0) > 0)
useClampedPage(page, {data: botPage, isPlaceholderData})
const newBotUsername = ref('')
const newBotName = ref('')
const createError = ref<string | null>(null)
const showCreateForm = ref(false)

const tokenPages = ref<Record<number, number>>({})
const tokenQueries = useQueries({queries: computed(() => bots.value.map(bot => botApiTokensQuery(
	bot.id,
	tokenPages.value[bot.id] ?? 1,
)))})
const tokenPageByBot = computed(() => Object.fromEntries(
	bots.value.map((bot, index) => [bot.id, tokenQueries.value[index]?.data]),
))
watch(tokenQueries, results => bots.value.forEach((bot, index) => {
	const result = results[index]
	if (!result) return
	const current = tokenPages.value[bot.id] ?? 1
	const clamped = clampPage(current, result)
	if (clamped !== current) tokenPages.value[bot.id] = clamped
}))
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
		delete tokenPages.value[bot.id]
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
	<div
		class="content loader-container"
		:class="{'is-loading': isPlaceholderData}"
	>
		<h2>{{ $t('user.settings.bots.title') }}</h2>
		<p>{{ $t('user.settings.bots.description') }}</p>

		<div
			v-if="!hasBots || showCreateForm"
			class="create-form"
		>
			<FormField
				:label="$t('user.auth.username')"
				:error="createError"
			>
				<template #default="{id}">
					<input
						:id="id"
						v-model="newBotUsername"
						class="input"
						placeholder="bot-myassistant"
					>
				</template>
			</FormField>
			<FormField :label="$t('admin.users.nameLabel')">
				<template #default="{id}">
					<input
						:id="id"
						v-model="newBotName"
						class="input"
						:placeholder="$t('user.settings.bots.namePlaceholder')"
					>
				</template>
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
					v-if="(tokenPageByBot[bot.id]?.items ?? []).length > 0"
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
								v-for="token in tokenPageByBot[bot.id]?.items ?? []"
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
				<PaginationEmit
					:total-pages="tokenPageByBot[bot.id]?.total_pages ?? 0"
					:current-page="tokenPages[bot.id] ?? 1"
					@pageChanged="(page: number) => tokenPages[bot.id] = page"
				/>
				<p
					v-if="bot.status !== STATUS_ACTIVE"
					class="help"
				>
					{{ $t('user.settings.bots.tokensInactive') }}
				</p>
				<ApiTokenForm
					v-else-if="showTokenForm[bot.id]"
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
			</div>
		</div>

		<Pagination
			:total-pages="totalPages"
			:current-page="page"
		/>

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
