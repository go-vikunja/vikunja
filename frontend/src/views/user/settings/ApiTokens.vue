<script setup lang="ts">
import ApiTokenService from '@/services/apiToken'
import {onMounted, ref} from 'vue'
import {useRoute} from 'vue-router'
import {formatDateSince, formatDisplayDate} from '@/helpers/time/formatDate'
import XButton from '@/components/input/Button.vue'
import BaseButton from '@/components/base/BaseButton.vue'
import {useI18n} from 'vue-i18n'
import Message from '@/components/misc/Message.vue'
import type {IApiToken} from '@/modelTypes/IApiToken'
import ApiTokenForm from '@/components/token/ApiTokenForm.vue'

const service = new ApiTokenService()
const tokens = ref<IApiToken[]>([])
const apiDocsUrl = window.API_URL + '/docs'
const showCreateForm = ref(false)
const tokenCreatedSuccessMessage = ref('')

const showDeleteModal = ref<boolean>(false)
const tokenToDelete = ref<IApiToken>()

const {t} = useI18n()

const route = useRoute()

onMounted(async () => {
	tokens.value = await service.getAll()

	// Apply query parameters if present
	const titleParam = Array.isArray(route.query.title) ? route.query.title[0] : route.query.title
	const scopesParam = Array.isArray(route.query.scopes) ? route.query.scopes[0] : route.query.scopes

	if (titleParam || scopesParam) {
		showCreateForm.value = true
	}
})

async function deleteToken() {
	await service.delete(tokenToDelete.value)
	showDeleteModal.value = false
	const index = tokens.value.findIndex(el => el.id === tokenToDelete.value.id)
	tokenToDelete.value = null
	if (index === -1) {
		return
	}
	tokens.value.splice(index, 1)
}

function formatPermissionTitle(title: string): string {
	return title.replaceAll('_', ' ')
}

function onTokenCreated(token: IApiToken) {
	tokenCreatedSuccessMessage.value = t('user.settings.apiTokens.tokenCreatedSuccess', {token: token.token})
	tokens.value.push(token)
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
								{{ v.map(formatPermissionTitle).join(', ') }}
								<br>
							</template>
						</td>
						<td>
							{{ formatDisplayDate(tk.expiresAt) }}
							<p
								v-if="tk.expiresAt < new Date()"
								class="has-text-danger"
							>
								{{ $t('user.settings.apiTokens.expired', {ago: formatDateSince(tk.expiresAt)}) }}
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
			:loading="service.loading"
			@created="onTokenCreated"
		/>

		<XButton
			v-else
			icon="plus"
			class="mbe-4"
			:loading="service.loading"
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
				<p>
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
