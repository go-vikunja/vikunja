<script setup lang="ts">
import {computed, onMounted, ref, shallowReactive} from 'vue'
import {useNow} from '@vueuse/core'
import {useI18n} from 'vue-i18n'
import ApiTokenService from '@/services/apiToken'
import {mcpInfo, type ConnectionSettings} from '@/client/generated'
import type {IApiToken} from '@/modelTypes/IApiToken'
import type {ApiTokenPreset, ApiTokenPresetGroups} from '@/modelTypes/IApiTokenSettings'
import ApiTokenForm from '@/components/token/ApiTokenForm.vue'
import McpClientGuide from '@/components/token/McpClientGuide.vue'
import XButton from '@/components/input/Button.vue'
import FormField from '@/components/input/FormField.vue'
import Message from '@/components/misc/Message.vue'
import {useCopyToClipboard} from '@/composables/useCopyToClipboard'
import {useTitle} from '@/composables/useTitle'
import {formatDateSince, formatDisplayDate} from '@/helpers/time/formatDate'
import {error} from '@/message'
import {MCP_HELP} from '@/urls'

defineOptions({name: 'McpSettings'})

const {t} = useI18n({useScope: 'global'})
useTitle(() => `${t('user.settings.mcp.title')} - ${t('user.settings.title')}`)
const service = shallowReactive(new ApiTokenService())
const info = ref<ConnectionSettings>()
const endpoint = computed(() => info.value?.endpoint ?? '')
const tokens = ref<IApiToken[]>([])
const loading = ref(true)
const loadFailed = ref(false)
const showCreateForm = ref(false)
const newToken = ref('')
const tokenToDelete = ref<IApiToken>()
const showDeleteModal = ref(false)
const copy = useCopyToClipboard()
const now = useNow({interval: 60_000})

function presetGroups(groups: Record<string, string | string[] | null> = {}): ApiTokenPresetGroups {
	return Object.fromEntries(Object.entries(groups).map(([group, permissions]) => [
		group,
		permissions === '*' ? '*' : Array.isArray(permissions) ? permissions : [],
	]))
}

const presets = computed<ApiTokenPreset[]>(() => info.value ? [
	{id: 'readOnly', groups: presetGroups(info.value.presets?.read_only)},
	{id: 'typed', groups: presetGroups(info.value.presets?.typed), label: t('user.settings.mcp.typedPreset')},
	{id: 'fullAccess', groups: presetGroups(info.value.presets?.full)},
] : [])
const initialScopes = computed(() => Object.entries(info.value?.presets?.typed ?? {})
	.flatMap(([group, permissions]) => (permissions ?? []).map(permission => `${group}:${permission}`))
	.join(','))

async function refreshTokens() {
	const allTokens = await service.getAll()
	for (let page = 2; page <= service.totalPages; page++) {
		allTokens.push(...await service.getAll(undefined, {}, page))
	}
	tokens.value = allTokens.filter(token => token.permissions.mcp?.includes('access'))
}

async function load() {
	loading.value = true
	loadFailed.value = false
	try {
		const [result] = await Promise.all([mcpInfo(), refreshTokens()])
		info.value = result.data
	} catch (e) {
		loadFailed.value = true
		error(e)
	} finally {
		loading.value = false
	}
}
onMounted(load)

function onTokenCreated(token: IApiToken) {
	newToken.value = token.token
	showCreateForm.value = false
}

async function done() {
	newToken.value = ''
	try {
		await refreshTokens()
	} catch (e) {
		error(e)
	}
}

async function deleteToken() {
	const token = tokenToDelete.value
	if (!token) return
	tokenToDelete.value = undefined
	showDeleteModal.value = false
	try {
		await service.delete(token)
		tokens.value = tokens.value.filter(({id}) => id !== token.id)
	} catch (e) {
		error(e)
	}
}
</script>

<template>
	<Card
		:title="t('user.settings.mcp.title')"
		:loading="loading"
	>
		<p>
			{{ t('user.settings.mcp.intro') }}
			<a
				:href="MCP_HELP"
				target="_blank"
				rel="noreferrer"
			>{{ t('user.settings.mcp.more') }}</a>
		</p>
		<XButton
			v-if="loadFailed"
			@click="load"
		>
			{{ t('sharing.retry') }}
		</XButton>
		<template v-if="info">
			<label
				class="label"
				for="mcp-endpoint"
			>{{ t('user.settings.mcp.endpoint') }}</label>
			<FormField
				id="mcp-endpoint"
				:model-value="endpoint"
				readonly
			>
				<template #addon>
					<XButton
						:shadow="false"
						icon="paste"
						:aria-label="t('misc.copy')"
						@click="copy(endpoint)"
					/>
				</template>
			</FormField>
			<template v-if="newToken">
				<Message
					variant="success"
					class="mbe-4"
				>
					{{ t('user.settings.mcp.created') }}<br>
					{{ t('user.settings.apiTokens.tokenCreatedNotSeeAgain') }}
				</Message>
				<McpClientGuide
					:endpoint="endpoint"
					:token="newToken"
				/>
				<XButton
					class="mbs-4"
					@click="done"
				>
					{{ t('task.attributes.done') }}
				</XButton>
			</template>
			<template v-else>
				<h5 class="mbs-5 mbe-4 has-text-weight-bold">
					{{ t('user.settings.mcp.tokens') }}
				</h5>
				<div
					v-if="tokens.length"
					class="has-horizontal-overflow"
				>
					<table class="table">
						<thead>
							<tr>
								<th>{{ t('user.settings.apiTokens.attributes.title') }}</th>
								<th>{{ t('user.settings.apiTokens.attributes.expiresAt') }}</th>
								<th>{{ t('misc.created') }}</th>
								<th class="has-text-end">
									{{ t('misc.actions') }}
								</th>
							</tr>
						</thead>
						<tbody>
							<tr
								v-for="token in tokens"
								:key="token.id"
								:class="{'mcp-token-expired': token.expiresAt < now}"
							>
								<td>{{ token.title }}</td>
								<td>
									{{ formatDisplayDate(token.expiresAt) }}
									<p
										v-if="token.expiresAt < now"
										class="has-text-danger"
									>
										{{ t('user.settings.apiTokens.expired', {ago: formatDateSince(token.expiresAt)}) }}
									</p>
								</td>
								<td>{{ formatDisplayDate(token.created) }}</td>
								<td class="has-text-end">
									<XButton
										variant="secondary"
										@click="() => {tokenToDelete = token; showDeleteModal = true}"
									>
										{{ t('misc.delete') }}
									</XButton>
								</td>
							</tr>
						</tbody>
					</table>
				</div>
				<p class="mbe-4">
					<RouterLink :to="{name: 'user.settings.apiTokens'}">
						{{ t('user.settings.mcp.manageTokens') }}
					</RouterLink>
				</p>
				<ApiTokenForm
					v-if="showCreateForm"
					:routes="info.routes ?? {}"
					:presets="presets"
					:locked-scopes="{mcp: ['access']}"
					:initial-title="t('user.settings.mcp.title')"
					:initial-scopes="initialScopes"
					@created="onTokenCreated"
					@cancel="showCreateForm = false"
				/>
				<XButton
					v-else
					icon="plus"
					@click="showCreateForm = true"
				>
					{{ t('user.settings.apiTokens.createAToken') }}
				</XButton>
			</template>
		</template>
		<Modal
			:enabled="showDeleteModal"
			@close="showDeleteModal = false"
			@submit="deleteToken"
		>
			<template #header>
				{{ t('user.settings.apiTokens.delete.header') }}
			</template>
			<template #text>
				<p v-if="tokenToDelete">
					{{ t('user.settings.apiTokens.delete.text1', {token: tokenToDelete.title}) }}<br>
					{{ t('user.settings.apiTokens.delete.text2') }}
				</p>
			</template>
		</Modal>
	</Card>
</template>

<style scoped lang="scss">
.mcp-token-expired {
	opacity: .6;
}
</style>
