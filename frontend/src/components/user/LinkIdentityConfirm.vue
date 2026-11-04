<template>
	<div>
		<Message
			v-if="errorMessage"
			variant="danger"
			class="mbe-4"
		>
			{{ errorMessage }}
		</Message>
		<form
			v-if="mode === 'confirm'"
			@submit.prevent="linkIdentity"
		>
			<h2 class="title">
				{{ $t('user.settings.linkedAccounts.confirmTitle', {provider: providerName}) }}
			</h2>
			<p class="mbe-4">
				{{ $t('user.settings.linkedAccounts.confirmText', {provider: providerName}) }}
			</p>
			<Message v-if="!identity">
				{{ $t('user.auth.authenticating') }}
			</Message>
			<p
				v-else
				class="mbe-4"
			>
				{{ $t('user.settings.linkedAccounts.confirmAccount') }}
				<strong>{{ identity.account }}</strong>
			</p>
			<ReauthFields
				v-if="identity"
				v-model:password="password"
				v-model:passcode="totpPasscode"
				password-id="linkIdentityPassword"
				passcode-id="linkIdentityPasscode"
				:show-passcode="needsTotp"
				focus
			/>
			<div class="buttons mbs-4">
				<XButton
					v-if="identity"
					type="submit"
					:loading="linkMutation.isPending.value || totpPending"
				>
					{{ $t('user.settings.linkedAccounts.confirm') }}
				</XButton>
				<XButton
					variant="tertiary"
					:disabled="linkMutation.isPending.value"
					@click="cancelLinking"
				>
					{{ $t('misc.cancel') }}
				</XButton>
			</div>
		</form>

		<XButton
			v-if="backLink"
			ref="backButton"
			class="mbs-2"
			:to="backLink.to"
		>
			{{ backLink.label }}
		</XButton>
	</div>
</template>

<script setup lang="ts">
import {computed, nextTick, onMounted, onBeforeUnmount, ref, watch} from 'vue'
import {useRouter} from 'vue-router'
import {useI18n} from 'vue-i18n'
import {useQuery} from '@tanstack/vue-query'
import type {IdentityPreview, VikunjaErrorModel} from '@/client/generated'
import {useLinkIdentityMutation, usePreviewIdentityMutation} from '@/client/queries/identities'
import {totpQuery} from '@/client/queries/totp'
import {isRequestContextAbort} from '@/client/requestContext'
import ReauthFields from '@/components/user/ReauthFields.vue'
import Message from '@/components/misc/Message.vue'
import {getErrorText} from '@/message'
import {endOpenIdFlow, getRedirectUrlFromCurrentFrontendPath} from '@/helpers/redirectToProvider'
import {useConfigStore} from '@/stores/config'
import {
	ERROR_CODE_TOO_MANY_REAUTH_ATTEMPTS,
	ERROR_CODE_TOTP_PASSCODE_USED,
	ERROR_CODE_TOTP_REQUIRED,
	ERROR_CODE_WRONG_USERNAME_OR_PASSWORD,
} from '@/constants/auth'

const props = defineProps<{
	providerKey: string,
	code: string,
	state: string,
	hasSession: boolean,
	initialError: string,
}>()

const {t} = useI18n({useScope: 'global'})
const router = useRouter()
const configStore = useConfigStore()
// eslint-disable-next-line vue/no-setup-props-reactivity-loss
const mode = ref(props.hasSession ? (props.initialError ? 'failed' : 'confirm') : 'noSession')
// eslint-disable-next-line vue/no-setup-props-reactivity-loss
const errorMessage = ref(props.initialError || (props.hasSession ? '' : t('user.settings.linkedAccounts.noSession')))
const password = ref('')
const totpPasscode = ref('')
const provider = computed(() => configStore.auth.openid_connect.providers?.find(p => p.key === props.providerKey))
const providerName = computed(() => provider.value?.name ?? props.providerKey)
const linkMutation = useLinkIdentityMutation()
const previewMutation = usePreviewIdentityMutation()
const identity = ref<IdentityPreview>()
let active = true
onBeforeUnmount(() => { active = false })
const totp = useQuery(computed(() => ({
	...totpQuery(),
	enabled: mode.value === 'confirm' && configStore.totp_enabled,
})))
const needsTotp = computed(() => totp.data.value?.enabled === true || totp.isError.value)
const totpPending = computed(() => configStore.totp_enabled && totp.isPending.value)
const backButton = ref<{$el: HTMLElement} | null>(null)
const backLink = computed(() => mode.value === 'confirm' ? null : {
	to: {name: mode.value === 'noSession' ? 'user.login' : 'user.settings.linked-accounts'},
	label: mode.value === 'noSession' ? t('user.auth.login') : t('user.settings.linkedAccounts.backToLinkedAccounts'),
})
watch(() => backLink.value?.to.name, focusBackButton)

async function focusBackButton() {
	await nextTick()
	backButton.value?.$el.focus()
}

// Rejected before consuming the prepared identity, so the same callback can retry.
const RETRYABLE_LINK_ERRORS = new Set([
	ERROR_CODE_WRONG_USERNAME_OR_PASSWORD,
	ERROR_CODE_TOTP_REQUIRED,
	ERROR_CODE_TOTP_PASSCODE_USED,
	ERROR_CODE_TOO_MANY_REAUTH_ATTEMPTS,
])

async function linkIdentity() {
	if (!identity.value || linkMutation.isPending.value || totpPending.value) return
	if (!provider.value) {
		errorMessage.value = t('user.auth.openIdGeneralError')
		return
	}

	errorMessage.value = ''
	try {
		await linkMutation.mutateAsync({
			provider: props.providerKey,
			code: props.code,
			redirect_url: getRedirectUrlFromCurrentFrontendPath(provider.value),
			password: password.value,
			totp_passcode: needsTotp.value ? totpPasscode.value : undefined,
		})
	} catch (e) {
		if (isRequestContextAbort(e)) return
		errorMessage.value = getErrorText(e)
		const code = (e as VikunjaErrorModel)?.code
		if (code === undefined || !RETRYABLE_LINK_ERRORS.has(code)) {
			endOpenIdFlow(props.state)
			mode.value = 'failed'
		}
		return
	}

	endOpenIdFlow(props.state)
	await router.replace({name: 'user.settings.linked-accounts'})
}

function cancelLinking() {
	active = false
	endOpenIdFlow(props.state)
	router.replace({name: 'user.settings.linked-accounts'})
}

onMounted(async () => {
	if (mode.value !== 'confirm') {
		endOpenIdFlow(props.state)
		focusBackButton()
		return
	}

	try {
		if (!provider.value) throw new Error(t('user.auth.openIdGeneralError'))
		const preview = await previewMutation.mutateAsync({
			provider: props.providerKey,
			code: props.code,
			redirect_url: getRedirectUrlFromCurrentFrontendPath(provider.value),
		})
		if (active) identity.value = preview
	} catch (e) {
		if (!active || isRequestContextAbort(e)) return
		errorMessage.value = getErrorText(e)
		endOpenIdFlow(props.state)
		mode.value = 'failed'
	}
})
</script>
