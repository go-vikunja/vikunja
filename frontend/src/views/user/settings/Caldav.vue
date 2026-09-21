<template>
	<Card
		v-if="configStore.caldav_enabled"
		:title="$t('user.settings.caldav.title')"
	>
		<p>
			{{ $t('user.settings.caldav.howTo') }}
		</p>
		<FormField
			v-model="caldavUrl"
			type="text"
			readonly
		>
			<template #addon>
				<XButton
					v-tooltip="$t('misc.copy')"
					:shadow="false"
					icon="paste"
					:aria-label="$t('misc.copy')"
					@click="copy(caldavUrl)"
				/>
			</template>
		</FormField>

		<h5 class="mbs-5 mbe-4 has-text-weight-bold">
			{{ $t('user.settings.caldav.tokens') }}
		</h5>

		<p>
			{{ isLocalUser ? $t('user.settings.caldav.tokensHowTo') : $t('user.settings.caldav.mustUseToken') }}
			<br>
			<i18n-t
				keypath="user.settings.caldav.usernameIs"
				scope="global"
			>
				<strong>{{ username }}</strong>
			</i18n-t>
		</p>

		<p class="mbs-2">
			<i18n-t
				keypath="user.settings.caldav.apiTokenHint"
				scope="global"
			>
				<template #link>
					<RouterLink :to="{name: 'user.settings.apiTokens'}">
						{{ $t('user.settings.apiTokens.title') }}
					</RouterLink>
				</template>
			</i18n-t>
		</p>

		<table
			v-if="tokens.length > 0"
			class="table"
		>
			<tr>
				<th>{{ $t('misc.id') }}</th>
				<th>{{ $t('misc.created') }}</th>
				<th class="has-text-end">
					{{ $t('misc.actions') }}
				</th>
			</tr>
			<tr
				v-for="tk in tokens"
				:key="tk.id"
			>
				<td>{{ tk.id }}</td>
				<td>{{ formatDateShort(tk.created) }}</td>
				<td class="has-text-end">
					<XButton
						variant="secondary"
						@click="deleteToken(tk)"
					>
						{{ $t('misc.delete') }}
					</XButton>
				</td>
			</tr>
		</table>

		<Message
			v-if="newToken"
			class="mbe-4"
		>
			{{ $t('user.settings.caldav.tokenCreated', {token: newToken.token}) }}<br>
			{{ $t('user.settings.caldav.wontSeeItAgain') }}
		</Message>

		<XButton
			icon="plus"
			class="mbe-4"
			:loading="createMutation.isPending.value"
			@click="createToken"
		>
			{{ $t('user.settings.caldav.createToken') }}
		</XButton>

		<p>
			<BaseButton
				:href="CALDAV_DOCS"
				target="_blank"
			>
				{{ $t('user.settings.caldav.more') }}
			</BaseButton>
		</p>
	</Card>
</template>

<script lang="ts" setup>
import {computed, ref} from 'vue'
import {useI18n} from 'vue-i18n'

import {CALDAV_DOCS} from '@/urls'
import {useTitle} from '@/composables/useTitle'
import {useCopyToClipboard} from '@/composables/useCopyToClipboard'
import BaseButton from '@/components/base/BaseButton.vue'
import Message from '@/components/misc/Message.vue'
import FormField from '@/components/input/FormField.vue'
import {useQuery} from '@tanstack/vue-query'
import {caldavTokensQuery, useCreateCaldavTokenMutation, useDeleteCaldavTokenMutation} from '@/client/queries/caldavTokens'
import { formatDateShort } from '@/helpers/time/formatDate'
import type {Token} from '@/client/generated'
import {useConfigStore} from '@/stores/config'
import {useAuthStore} from '@/stores/auth'

const copy = useCopyToClipboard()

const {t} = useI18n({useScope: 'global'})
useTitle(() => `${t('user.settings.caldav.title')} - ${t('user.settings.title')}`)

const authStore = useAuthStore()
const configStore = useConfigStore()
const caldav_enabled = computed(() => configStore.caldav_enabled)
const tokenQuery = useQuery(computed(() => ({...caldavTokensQuery(), enabled: caldav_enabled.value})))
const tokens = computed(() => tokenQuery.data.value ?? [])
const createMutation = useCreateCaldavTokenMutation()
const deleteMutation = useDeleteCaldavTokenMutation()
const newToken = ref<Token>()

async function createToken() {
	try {
		newToken.value = await createMutation.mutateAsync()
		// Detaching the observer lets gcTime: 0 drop the cached token.
		createMutation.reset()
	} catch { return }
}

async function deleteToken(token: Token) {
	if (!token.id) return
	try {
		await deleteMutation.mutateAsync(token.id)
		if (newToken.value?.id === token.id) newToken.value = undefined
	} catch { return }
}

const username = computed(() => authStore.info?.username)
const caldavUrl = computed(() => `${configStore.apiBase}/dav/principals/${username.value}/`)
const isLocalUser = computed(() => authStore.info?.is_local_user)
</script>
