<template>
	<div>
		<Message
			v-if="errorMessage"
			variant="danger"
		>
			{{ errorMessage }}
		</Message>
		<Message
			v-if="errorMessageFromQuery"
			variant="danger"
			class="mbs-2"
		>
			{{ errorMessageFromQuery }}
		</Message>
		<Message v-if="loading && !needsTotp">
			{{ $t('user.auth.authenticating') }}
		</Message>

		<form
			v-if="needsTotp"
			@submit.prevent="submitTotpAndRestart"
		>
			<Message class="mbe-2">
				{{ $t('user.auth.openIdTotpRequired') }}
			</Message>
			<FormField
				id="openIdTotpPasscode"
				ref="totpInput"
				v-model="totpPasscode"
				v-focus
				:label="$t('user.auth.totpTitle')"
				autocomplete="one-time-code"
				:placeholder="$t('user.auth.totpPlaceholder')"
				required
				type="text"
				inputmode="numeric"
			/>
			<XButton
				:loading="loading"
				:disabled="!totpPasscode"
				class="mbs-2"
				@click="submitTotpAndRestart"
			>
				{{ $t('user.auth.openIdTotpSubmit') }}
			</XButton>
		</form>
	</div>
</template>


<script setup lang="ts">
import {ref, computed, onMounted} from 'vue'
import {useRoute} from 'vue-router'
import {useI18n} from 'vue-i18n'

import {getErrorText} from '@/message'
import Message from '@/components/misc/Message.vue'
import FormField from '@/components/input/FormField.vue'
import {useRedirectToLastVisited} from '@/composables/useRedirectToLastVisited'
import {redirectToProvider} from '@/helpers/redirectToProvider'

import {useAuthStore} from '@/stores/auth'
import {useConfigStore} from '@/stores/config'
import type {IProvider} from '@/types/IProvider'

defineOptions({name: 'Auth'})

const {t} = useI18n({useScope: 'global'})

const route = useRoute()
const {redirectIfSaved} = useRedirectToLastVisited()

const authStore = useAuthStore()
const configStore = useConfigStore()

const loading = computed(() => authStore.isLoading)
const errorMessage = ref('')
const errorMessageFromQuery = computed(() => route.query.error)

const needsTotp = ref(false)
const totpPasscode = ref('')

function pendingTotpKey(provider: string): string {
	return `openid_pending_totp_${provider}`
}

function findProvider(providerKey: string): IProvider | undefined {
	return configStore.auth.openidConnect.providers?.find((p: IProvider) => p.key === providerKey)
}

async function authenticateWithCode() {
	errorMessage.value = ''

	const providerKey = route.params.provider as string

	if (typeof route.query.error !== 'undefined') {
		sessionStorage.removeItem(pendingTotpKey(providerKey))
		errorMessage.value = typeof route.query.message !== 'undefined'
			? route.query.message as string
			: t('user.auth.openIdGeneralError')
		return
	}

	const state = localStorage.getItem('state')
	if (typeof route.query.state === 'undefined' || route.query.state !== state) {
		sessionStorage.removeItem(pendingTotpKey(providerKey))
		errorMessage.value = t('user.auth.openIdStateError')
		return
	}

	// sessionStorage (not localStorage): per-tab, cleared on tab close.
	const pendingPasscode = sessionStorage.getItem(pendingTotpKey(providerKey)) ?? undefined
	if (pendingPasscode) {
		sessionStorage.removeItem(pendingTotpKey(providerKey))
	}

	try {
		await authStore.openIdAuth({
			provider: providerKey,
			code: route.query.code as string,
			totpPasscode: pendingPasscode,
		})

		redirectIfSaved()
	} catch (e) {
		const err = e as {response?: {data?: {code?: number}}}
		if (err?.response?.data?.code === 1017) {
			needsTotp.value = true
			return
		}
		errorMessage.value = getErrorText(e)
	}
}

async function submitTotpAndRestart() {
	if (!totpPasscode.value) {
		return
	}

	const providerKey = route.params.provider as string
	const provider = findProvider(providerKey)
	if (!provider) {
		errorMessage.value = t('user.auth.openIdGeneralError')
		return
	}

	sessionStorage.setItem(pendingTotpKey(providerKey), totpPasscode.value)
	// The auth code is single-use; restart the OIDC flow so the next callback reads the stashed passcode.
	redirectToProvider(provider)
}

onMounted(() => authenticateWithCode())
</script>
