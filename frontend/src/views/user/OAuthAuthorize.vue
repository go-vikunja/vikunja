<template>
	<div>
		<Message
			v-if="errorMessage"
			variant="danger"
		>
			{{ errorMessage }}
		</Message>
		<Message v-if="redirectedToApp">
			{{ $t('user.auth.oauthRedirectedToApp') }}
		</Message>
		<Message v-else-if="loading">
			{{ $t('user.auth.authenticating') }}
		</Message>
	</div>
</template>

<script setup lang="ts">
import {ref, onMounted} from 'vue'
import {useRoute} from 'vue-router'
import {useI18n} from 'vue-i18n'

import {getErrorText} from '@/message'
import Message from '@/components/misc/Message.vue'
import {oauthAuthorize} from '@/client/generated'

defineOptions({name: 'OAuthAuthorize'})

const {t} = useI18n({useScope: 'global'})
const route = useRoute()

const loading = ref(true)
const errorMessage = ref('')
const redirectedToApp = ref(false)

const requiredParams = [
	'response_type',
	'client_id',
	'redirect_uri',
	'code_challenge',
	'code_challenge_method',
] as const

const q = (v: unknown) => typeof v === 'string' ? v : undefined

async function authorize() {
	// Validate required query parameters
	const missing = requiredParams.filter(p => !route.query[p])
	if (missing.length > 0) {
		errorMessage.value = t('user.auth.oauthMissingParams', {params: missing.join(', ')})
		loading.value = false
		return
	}

	try {
		const response = await oauthAuthorize({body: {
			response_type: q(route.query.response_type),
			client_id: q(route.query.client_id),
			redirect_uri: q(route.query.redirect_uri),
			state: q(route.query.state),
			code_challenge: q(route.query.code_challenge),
			code_challenge_method: q(route.query.code_challenge_method),
		}})

		const {code, redirect_uri, state} = response.data

		if (!redirect_uri || !code) throw new Error('Authorization response is incomplete')
		const redirectUrl = new URL(redirect_uri)
		redirectUrl.searchParams.set('code', code)
		if (state) {
			redirectUrl.searchParams.set('state', state)
		}

		redirectedToApp.value = true
		loading.value = false

		window.location.href = redirectUrl.toString()
	} catch (e) {
		errorMessage.value = getErrorText(e)
		loading.value = false
	}
}

onMounted(() => authorize())
</script>
