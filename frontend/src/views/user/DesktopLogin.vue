<template>
	<div>
		<Message
			v-if="errorMessage"
			variant="danger"
			class="mbe-4"
		>
			{{ errorMessage }}
		</Message>
		<Message
			v-if="waitingForAuth"
			class="mbe-4"
		>
			{{ $t('user.auth.desktopWaitingForAuth') }}
		</Message>

		<template v-if="hasStoredServer">
			<XButton
				:loading="waitingForAuth"
				class="is-fullwidth"
				@click="loginWithServer(window.API_URL)"
			>
				{{ $t('user.auth.login') }}
			</XButton>
		</template>
		<template v-else-if="showCustomServerInput">
			<p class="mbe-4">
				{{ $t('user.auth.desktopCustomServerDescription') }}
			</p>
			<ApiConfig
				:configure-open="true"
				@foundApi="loginWithServer"
			/>
			<div class="has-text-centered mbs-2">
				<a
					role="button"
					@click="showCustomServerInput = false"
				>
					{{ $t('misc.cancel') }}
				</a>
			</div>
		</template>
		<template v-else>
			<XButton
				:loading="waitingForAuth"
				class="is-fullwidth mbe-2"
				@click="loginWithServer('https://app.vikunja.cloud')"
			>
				Vikunja Cloud
			</XButton>
			<XButton
				:loading="waitingForAuth"
				variant="secondary"
				class="is-fullwidth mbe-2"
				@click="loginWithServer('https://try.vikunja.io')"
			>
				{{ $t('user.auth.desktopTryDemo') }}
			</XButton>
			<XButton
				variant="secondary"
				class="is-fullwidth"
				@click="showCustomServerInput = true"
			>
				{{ $t('user.auth.desktopCustomServer') }}
			</XButton>
		</template>
	</div>
</template>

<script setup lang="ts">
import {ref} from 'vue'
import {useI18n} from 'vue-i18n'

import Message from '@/components/misc/Message.vue'
import ApiConfig from '@/components/misc/ApiConfig.vue'

import {getErrorText} from '@/message'
import {startDesktopOAuthLogin, listenForDesktopOAuthTokens, listenForDesktopOAuthError} from '@/helpers/desktopAuth'
import {checkAndSetApiUrl} from '@/helpers/checkAndSetApiUrl'
import {useRedirectToLastVisited} from '@/composables/useRedirectToLastVisited'

import {useAuthStore} from '@/stores/auth'

const {t} = useI18n({useScope: 'global'})
const authStore = useAuthStore()
const {redirectIfSaved} = useRedirectToLastVisited()

const waitingForAuth = ref(false)
const errorMessage = ref('')
const hasStoredServer = localStorage.getItem('API_URL') !== null
const showCustomServerInput = ref(false)

listenForDesktopOAuthTokens(async (tokens) => {
	waitingForAuth.value = false
	try {
		await authStore.handleDesktopOAuthTokens(tokens)
		redirectIfSaved()
	} catch (e) {
		errorMessage.value = getErrorText(e)
	}
})

listenForDesktopOAuthError((error) => {
	waitingForAuth.value = false
	errorMessage.value = t('user.auth.desktopOAuthError', {error})
})

async function loginWithServer(serverUrl: string) {
	errorMessage.value = ''
	waitingForAuth.value = true
	try {
		await checkAndSetApiUrl(serverUrl)
		await startDesktopOAuthLogin(window.API_URL)
	} catch (e) {
		waitingForAuth.value = false
		errorMessage.value = getErrorText(e)
	}
}
</script>
