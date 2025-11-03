<template>
	<Card
		v-if="caldavEnabled"
		:title="$t('user.settings.caldav.title')"
	>
		<p>
			{{ $t('user.settings.caldav.howTo') }}
		</p>
		<div class="field has-addons no-input-mobile">
			<div class="control is-expanded">
				<input
					v-model="caldavUrl"
					type="text"
					class="input"
					readonly
				>
			</div>
			<div class="control">
				<XButton
					v-tooltip="$t('misc.copy')"
					:shadow="false"
					icon="paste"
					@click="copy(caldavUrl)"
				/>
			</div>
		</div>

		<h5 class="mbs-5 mbe-4 has-text-weight-bold">
			{{ $t('user.settings.caldav.tokens') }}
		</h5>

		<p>
			{{ isLocalUser ? $t('user.settings.caldav.tokensHowTo') : $t('user.settings.caldav.mustUseToken') }}
			<template v-if="!isLocalUser">
				<br>
				<i18n-t
					keypath="user.settings.caldav.usernameIs"
					scope="global"
				>
					<strong>{{ username }}</strong>
				</i18n-t>
			</template>
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
			:loading="service.loading"
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
import {computed, ref, shallowReactive} from 'vue'
import {useI18n} from 'vue-i18n'

import {CALDAV_DOCS} from '@/urls'
import {useTitle} from '@/composables/useTitle'
import {useCopyToClipboard} from '@/composables/useCopyToClipboard'
import {success} from '@/message'
import BaseButton from '@/components/base/BaseButton.vue'
import Message from '@/components/misc/Message.vue'
import CaldavTokenService from '@/services/caldavToken'
import { formatDateShort } from '@/helpers/time/formatDate'
import type {ICaldavToken} from '@/modelTypes/ICaldavToken'
import {useConfigStore} from '@/stores/config'
import {useAuthStore} from '@/stores/auth'

const copy = useCopyToClipboard()

const {t} = useI18n({useScope: 'global'})
useTitle(() => `${t('user.settings.caldav.title')} - ${t('user.settings.title')}`)

const service = shallowReactive(new CaldavTokenService())
const tokens = ref<ICaldavToken[]>([])

service.getAll().then((result: ICaldavToken[]) => {
	tokens.value = result
})

const newToken = ref<ICaldavToken>()
async function createToken() {
	newToken.value = await service.create({}) as ICaldavToken
	tokens.value.push(newToken.value)
}

async function deleteToken(token: ICaldavToken) {
	const r = await service.delete(token)
	tokens.value = tokens.value.filter(({id}) => id !== token.id)
	success(r)
}

const authStore = useAuthStore()
const configStore = useConfigStore()
const username = computed(() => authStore.info?.username)
const caldavUrl = computed(() => `${configStore.apiBase}/dav/principals/${username.value}/`)
const caldavEnabled = computed(() => configStore.caldavEnabled)
const isLocalUser = computed(() => authStore.info?.isLocalUser)
</script>
