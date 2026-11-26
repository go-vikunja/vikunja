<template>
	<div>
		<Message
			v-if="errorMessage"
			variant="danger"
			class="mbe-4"
		>
			{{ errorMessage }}
		</Message>
		<Message v-if="phase.kind === 'redirected'">
			{{ $t('user.auth.oauthRedirectedToApp') }}
		</Message>
		<Message v-else-if="phase.kind === 'loading'">
			{{ $t('user.auth.authenticating') }}
		</Message>
		<OAuthConsent
			v-else-if="phase.kind === 'consent' && authStore.info"
			:client="phase.preview.client"
			:redirect-uri="phase.preview.redirectUri"
			:requested="phase.preview.requested"
			:user="authStore.info"
			:pending="approving"
			@approve="approve"
			@deny="deny"
		/>
	</div>
</template>

<script setup lang="ts">
import {ref, onMounted} from 'vue'
import {useRoute} from 'vue-router'
import {useI18n} from 'vue-i18n'

import {getErrorText} from '@/message'
import Message from '@/components/misc/Message.vue'
import OAuthConsent, {type AccessLevel} from '@/components/user/OAuthConsent.vue'
import {
	oauthAuthorize,
	oauthAuthorizePreview,
	type AuthorizeClient,
} from '@/client/generated'
import {isRequestContextAbort} from '@/client/requestContext'
import {useAuthStore} from '@/stores/auth'

defineOptions({name: 'OAuthAuthorize'})

interface ConsentPreview {
	client: AuthorizeClient
	redirectUri: string
	requested: AccessLevel
}

type Phase =
	| {kind: 'loading'}
	| {kind: 'failed'}
	| {kind: 'consent', preview: ConsentPreview}
	| {kind: 'redirected'}

const {t} = useI18n({useScope: 'global'})
const route = useRoute()
const authStore = useAuthStore()

const phase = ref<Phase>({kind: 'loading'})
const errorMessage = ref('')
const approving = ref(false)

const requiredParams = [
	'response_type',
	'client_id',
	'redirect_uri',
	'code_challenge',
	'code_challenge_method',
] as const

const SCOPE_FOR_LEVEL: Record<AccessLevel, string> = {
	read: 'mcp:read',
	write: 'mcp:read mcp:write',
}

const q = (v: unknown) => typeof v === 'string' ? v : undefined

function authorizeRequest() {
	return {
		response_type: q(route.query.response_type),
		client_id: q(route.query.client_id),
		redirect_uri: q(route.query.redirect_uri),
		state: q(route.query.state),
		code_challenge: q(route.query.code_challenge),
		code_challenge_method: q(route.query.code_challenge_method),
		scope: q(route.query.scope),
		resource: q(route.query.resource),
	}
}

function fail(e: unknown) {
	if (isRequestContextAbort(e)) return
	errorMessage.value = getErrorText(e)
	phase.value = {kind: 'failed'}
}

function redirectTo(redirectUri: string, params: Record<string, string | undefined>) {
	const url = new URL(redirectUri)
	for (const [key, value] of Object.entries(params)) {
		if (value) url.searchParams.set(key, value)
	}
	phase.value = {kind: 'redirected'}
	window.location.href = url.toString()
}

async function authorize(scope?: string) {
	const response = await oauthAuthorize({body: {
		...authorizeRequest(),
		scope,
	}})
	const {code, redirect_uri, state} = response.data
	if (!redirect_uri || !code) throw new Error('Authorization response is incomplete')
	redirectTo(redirect_uri, {
		code,
		state,
	})
}

async function start() {
	const {data} = await oauthAuthorizePreview({body: authorizeRequest()})
	// First-party apps rely on being approved without a prompt. The server decides which requests
	// those are: a URL client is scoped even without scope or resource.
	if (!data.scoped) {
		await authorize()
		return
	}
	if (!data.client || !data.redirect_uri || !data.scopes?.length) {
		throw new Error('Authorization preview is incomplete')
	}
	phase.value = {
		kind: 'consent',
		preview: {
			client: data.client,
			redirectUri: data.redirect_uri,
			requested: data.scopes.includes('mcp:write') ? 'write' : 'read',
		},
	}
}

async function approve(level: AccessLevel) {
	if (approving.value) return
	approving.value = true
	errorMessage.value = ''
	try {
		await authorize(SCOPE_FOR_LEVEL[level])
	} catch (e) {
		if (isRequestContextAbort(e)) return
		errorMessage.value = getErrorText(e)
	} finally {
		approving.value = false
	}
}

function deny() {
	if (phase.value.kind !== 'consent') return
	// RFC 6749 §4.1.2.1; only ever to the redirect URI the server validated.
	redirectTo(phase.value.preview.redirectUri, {
		error: 'access_denied',
		state: q(route.query.state),
	})
}

function isFramed() {
	try {
		return window.self !== window.top
	} catch {
		// Reading a cross-origin top can throw.
		return true
	}
}

onMounted(async () => {
	// Clickjacking: a framing site could overlay the approve button.
	if (isFramed()) {
		errorMessage.value = t('user.auth.oauthFramed')
		phase.value = {kind: 'failed'}
		return
	}

	const missing = requiredParams.filter(p => !route.query[p])
	if (missing.length > 0) {
		errorMessage.value = t('user.auth.oauthMissingParams', {params: missing.join(', ')})
		phase.value = {kind: 'failed'}
		return
	}

	try {
		await start()
	} catch (e) {
		fail(e)
	}
})
</script>
