<template>
	<card v-if="caldavEnabled" :title="$t('user.settings.caldav.title')">
		<p>
			{{ $t('user.settings.caldav.howTo') }}
		</p>
		<div class="field has-addons no-input-mobile">
			<div class="control is-expanded">
				<input type="text" v-model="caldavUrl" class="input" readonly/>
			</div>
			<div class="control">
				<x-button
					@click="copy(caldavUrl)"
					:shadow="false"
					v-tooltip="$t('misc.copy')"
					icon="paste"
				/>
			</div>
		</div>

		<p class="my-4 has-text-weight-bold">
			{{ $t('user.settings.caldav.tokens') }}
		</p>

		<p>
			{{ isLocalUser ? $t('user.settings.caldav.tokensHowTo') : $t('user.settings.caldav.mustUseToken') }}
			<template v-if="!isLocalUser">
				<br/>
				<i18n-t keypath="user.settings.caldav.usernameIs">
					<strong>{{ username }}</strong>
				</i18n-t>
			</template>
		</p>

		<table class="table" v-if="tokens.length > 0">
			<tr>
				<th>{{ $t('misc.id') }}</th>
				<th>{{ $t('misc.created') }}</th>
				<th class="has-text-right">{{ $t('misc.actions') }}</th>
			</tr>
			<tr v-for="tk in tokens">
				<td>{{ tk.id }}</td>
				<td>{{ formatDateShort(tk.created) }}</td>
				<td class="has-text-right">
					<x-button type="secondary" @click="deleteToken(tk)">
						{{ $t('misc.delete') }}
					</x-button>
				</td>
			</tr>
		</table>

		<Message v-if="newToken" class="mb-4">
			{{ $t('user.settings.caldav.tokenCreated', {token: newToken.token}) }}<br/>
			{{ $t('user.settings.caldav.wontSeeItAgain') }}
		</Message>

		<x-button icon="plus" class="mb-4" @click="createToken" :loading="service.loading">
			{{ $t('user.settings.caldav.createToken') }}
		</x-button>

		<p>
			<a :href="CALDAV_DOCS" rel="noreferrer noopener nofollow" target="_blank">
				{{ $t('user.settings.caldav.more') }}
			</a>
		</p>
	</card>
</template>

<script lang="ts" setup>
import copy from 'copy-to-clipboard'
import {computed, ref} from 'vue'
import {useI18n} from 'vue-i18n'
import {useStore} from 'vuex'

import {CALDAV_DOCS} from '@/urls'
import {useTitle} from '@/composables/useTitle'
import {success} from '@/message'
import Message from '../../../components/misc/message'
import CaldavTokenService from '../../../services/caldavToken'

const store = useStore()
const {t} = useI18n()

useTitle(() => `${t('user.settings.caldav.title')} - ${t('user.settings.title')}`)

const caldavUrl = computed(() => `${store.getters['config/apiBase']}/dav/principals/${store.state.auth.info.username}/`)
const caldavEnabled = computed(() => store.state.config.caldavEnabled)
const isLocalUser = computed(() => store.state.auth.info?.isLocalUser)
const username = computed(() => store.state.auth.info?.username)

const service = new CaldavTokenService()
const tokens = ref([])
service.getAll()
	.then(r => {
		tokens.value = r
	})

const newToken = ref(null)
const createToken = () => {
	service.create({})
		.then(r => {
			tokens.value.push(r)
			newToken.value = r
		})
}

const deleteToken = token => {
	service.delete(token)
		.then(r => {
			success(r)
			for (const i in tokens.value) {
				if (tokens.value[i].id === token.id) {
					tokens.value.splice(i, 1)
				}
			}
		})
}
</script>
