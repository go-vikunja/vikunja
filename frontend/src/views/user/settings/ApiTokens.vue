<script setup lang="ts">
import {useQuery} from '@tanstack/vue-query'
import {apiTokensQuery, useDeleteApiTokenMutation} from '@/client/queries/apiTokens'
import {parseDateOrNull} from '@/helpers/parseDateOrNull'
import {computed, onMounted, ref} from 'vue'
import {useRoute} from 'vue-router'
import {formatDateSince, formatDisplayDate} from '@/helpers/time/formatDate'
import XButton from '@/components/input/Button.vue'
import BaseButton from '@/components/base/BaseButton.vue'
import {useI18n} from 'vue-i18n'
import Message from '@/components/misc/Message.vue'
import type {ApiToken as IApiToken} from '@/client/generated'
import ApiTokenForm from '@/components/token/ApiTokenForm.vue'

const {data, isFetching} = useQuery(apiTokensQuery())
const tokens = computed(() => data.value ?? [])
const deleteMutation = useDeleteApiTokenMutation()
const apiDocsUrl = window.API_URL + '/docs'
const showCreateForm = ref(false)
const tokenCreatedSuccessMessage = ref('')

const showDeleteModal = ref<boolean>(false)
const tokenToDelete = ref<IApiToken>()

const {t} = useI18n()

const route = useRoute()

const initialTitle = ref('')
const initialScopes = ref('')

onMounted(() => {
	// Apply query parameters if present
	const titleParam = Array.isArray(route.query.title) ? route.query.title[0] : route.query.title
	const scopesParam = Array.isArray(route.query.scopes) ? route.query.scopes[0] : route.query.scopes

	if (titleParam) {
		initialTitle.value = titleParam
	}
	if (scopesParam) {
		initialScopes.value = scopesParam
	}
	if (titleParam || scopesParam) {
		showCreateForm.value = true
	}
})

async function deleteToken() {
	const token = tokenToDelete.value
	if (!token?.id) {
		return
	}
	tokenToDelete.value = undefined
	showDeleteModal.value = false
	try { await deleteMutation.mutateAsync(token.id) } catch { /* Mutation reports the error. */ }
}

function formatPermissionTitle(title: string): string {
	return title.replaceAll('_', ' ')
}

function onTokenCreated(token: IApiToken) {
	tokenCreatedSuccessMessage.value = t('user.settings.apiTokens.tokenCreatedSuccess', {token: token.token})
	showCreateForm.value = false
}
</script>

<template>
	<Card :title="$t('user.settings.apiTokens.title')">
		<Message
			v-if="tokenCreatedSuccessMessage !== ''"
			class="has-text-centered mbe-4"
		>
			{{ tokenCreatedSuccessMessage }}<br>
			{{ $t('user.settings.apiTokens.tokenCreatedNotSeeAgain') }}
		</Message>

		<p>
			{{ $t('user.settings.apiTokens.general') }}
			<BaseButton :href="apiDocsUrl">
				{{ $t('user.settings.apiTokens.apiDocs') }}
			</BaseButton>
			.
		</p>

		<div
			v-if="tokens.length > 0"
			class="has-horizontal-overflow"
		>
			<table class="table">
				<thead>
					<tr>
						<th>{{ $t('misc.id') }}</th>
						<th>{{ $t('user.settings.apiTokens.attributes.title') }}</th>
						<th>{{ $t('user.settings.apiTokens.attributes.permissions') }}</th>
						<th>{{ $t('user.settings.apiTokens.attributes.expiresAt') }}</th>
						<th>{{ $t('misc.created') }}</th>
						<th class="has-text-end">
							{{ $t('misc.actions') }}
						</th>
					</tr>
				</thead>
				<tbody>
					<tr
						v-for="tk in tokens"
						:key="tk.id"
					>
						<td>{{ tk.id }}</td>
						<td>{{ tk.title }}</td>
						<td class="is-capitalized">
							<template
								v-for="(v, p) in tk.permissions"
								:key="'permission-' + p"
							>
								<strong>{{ formatPermissionTitle(p) }}:</strong>
								{{ (v ?? []).map(formatPermissionTitle).join(', ') }}
								<br>
							</template>
						</td>
						<td>
							{{ formatDisplayDate(tk.expires_at) }}
							<p
								v-if="(parseDateOrNull(tk.expires_at)?.getTime() ?? Infinity) < Date.now()"
								class="has-text-danger"
							>
								{{ $t('user.settings.apiTokens.expired', {ago: formatDateSince(tk.expires_at)}) }}
							</p>
						</td>
						<td>{{ formatDisplayDate(tk.created) }}</td>
						<td class="has-text-end">
							<XButton
								variant="secondary"
								@click="() => {tokenToDelete = tk; showDeleteModal = true}"
							>
								{{ $t('misc.delete') }}
							</XButton>
						</td>
					</tr>
				</tbody>
			</table>
		</div>

		<ApiTokenForm
			v-if="showCreateForm"
			:initial-title="initialTitle"
			:initial-scopes="initialScopes"
			@created="onTokenCreated"
			@cancel="showCreateForm = false"
		/>

		<XButton
			v-else
			icon="plus"
			class="mbe-4"
			:loading="isFetching || deleteMutation.isPending.value"
			@click="() => showCreateForm = true"
		>
			{{ $t('user.settings.apiTokens.createAToken') }}
		</XButton>

		<Modal
			:enabled="showDeleteModal"
			@close="showDeleteModal = false"
			@submit="deleteToken()"
		>
			<template #header>
				{{ $t('user.settings.apiTokens.delete.header') }}
			</template>

			<template #text>
				<p v-if="tokenToDelete">
					{{ $t('user.settings.apiTokens.delete.text1', {token: tokenToDelete.title}) }}<br>
					{{ $t('user.settings.apiTokens.delete.text2') }}
				</p>
			</template>
		</Modal>
	</Card>
</template>

<style lang="scss" scoped>
.preset-buttons {
	margin-block-start: 1rem;
}
</style>
