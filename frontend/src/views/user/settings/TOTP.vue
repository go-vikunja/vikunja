<template>
	<Card
		v-if="configStore.totp_enabled && isLocalUser"
		:title="$t('user.settings.totp.title')"
	>
		<template v-if="status.isSuccess.value && !totp.enabled && !totp.secret">
			<ReauthFields
				v-model:password="totpPassword"
				password-id="totpEnrollPassword"
				:password-label="$t('user.settings.totp.enterPassword')"
				@submit="totpEnroll"
			/>
			<XButton
				:loading="enrollMutation.isPending.value"
				@click="totpEnroll"
			>
				{{ $t('user.settings.totp.enroll') }}
			</XButton>
		</template>
		<template v-else-if="isEnrolling">
			<p>
				{{ $t('user.settings.totp.finishSetupPart1') }}
				<strong>{{ totp.secret }}</strong><br>
				{{ $t('user.settings.totp.finishSetupPart2') }}
			</p>
			<p>
				{{ $t('user.settings.totp.scanQR') }}<br>
				<img
					:src="totpQR"
					alt=""
				>
			</p>
			<p>
				{{ $t('user.settings.totp.confirmNotice') }}
			</p>
			<ReauthFields
				v-model:password="totpPassword"
				v-model:passcode="totpConfirmPasscode"
				password-id="totpConfirmPassword"
				passcode-id="totpConfirmPasscode"
				:password-label="$t('user.settings.totp.enterPassword')"
				show-passcode
				@submit="totpConfirm"
			/>
			<XButton
				:loading="enableMutation.isPending.value"
				@click="totpConfirm"
			>
				{{ $t('misc.confirm') }}
			</XButton>
		</template>
		<template v-else-if="totp.enabled">
			<p>
				{{ $t('user.settings.totp.setupSuccess') }}
			</p>
			<p v-if="!totpDisableForm">
				<XButton
					danger
					@click="totpDisableForm = true"
				>
					{{ $t('misc.disable') }}
				</XButton>
			</p>
			<div v-if="totpDisableForm">
				<ReauthFields
					v-model:password="totpDisablePassword"
					v-model:passcode="totpDisablePasscode"
					password-id="currentPassword"
					passcode-id="totpDisablePasscode"
					:password-label="$t('user.settings.totp.enterPassword')"
					show-passcode
					focus
					@submit="totpDisable"
				/>
				<XButton
					danger
					:loading="disableMutation.isPending.value"
					@click="totpDisable"
				>
					{{ $t('user.settings.totp.disable') }}
				</XButton>
				<XButton
					variant="tertiary"
					class="mis-2"
					@click="closeDisableForm()"
				>
					{{ $t('misc.cancel') }}
				</XButton>
			</div>
		</template>
	</Card>
</template>


<script lang="ts" setup>
import {computed, ref} from 'vue'
import {useI18n} from 'vue-i18n'

import {useQuery} from '@tanstack/vue-query'
import {useObjectUrl} from '@vueuse/core'
import {
	totpQuery,
	totpQrQuery,
	useEnrollTotpMutation,
	useEnableTotpMutation,
	useDisableTotpMutation,
} from '@/client/queries/totp'
import ReauthFields from '@/components/user/ReauthFields.vue'

import {useTitle} from '@/composables/useTitle'
import {useConfigStore} from '@/stores/config'
import {useAuthStore} from '@/stores/auth'

defineOptions({name: 'UserSettingsTotp'})

const {t} = useI18n({useScope: 'global'})
useTitle(() => `${t('user.settings.totp.title')} - ${t('user.settings.title')}`)

const totpConfirmPasscode = ref('')
const totpPassword = ref('')
const totpDisableForm = ref(false)
const totpDisablePassword = ref('')
const totpDisablePasscode = ref('')

const configStore = useConfigStore()
const authStore = useAuthStore()
const isLocalUser = computed(() => authStore.info?.is_local_user)

const status = useQuery(computed(() => ({
	...totpQuery(),
	enabled: Boolean(configStore.totp_enabled && isLocalUser.value),
})))
const totp = computed(() => status.data.value ?? {})
const isEnrolling = computed(() => Boolean(totp.value.secret) && !totp.value.enabled)
const qr = useQuery(computed(() => ({...totpQrQuery(), enabled: isEnrolling.value})))
const totpQR = useObjectUrl(computed(() => isEnrolling.value ? qr.data.value : undefined))
const enrollMutation = useEnrollTotpMutation()
const enableMutation = useEnableTotpMutation()
const disableMutation = useDisableTotpMutation()

function totpEnroll() {
	if (enrollMutation.isPending.value) return
	enrollMutation.mutate(totpPassword.value)
}

async function totpConfirm() {
	if (enableMutation.isPending.value) return
	try {
		await enableMutation.mutateAsync({
			passcode: totpConfirmPasscode.value,
			password: totpPassword.value,
		})
		totpPassword.value = ''
		await authStore.logout()
	} catch { return }
}

function closeDisableForm() {
	totpDisableForm.value = false
	totpDisablePassword.value = ''
	totpDisablePasscode.value = ''
}

async function totpDisable() {
	if (disableMutation.isPending.value) return
	try {
		await disableMutation.mutateAsync({
			password: totpDisablePassword.value,
			totp_passcode: totpDisablePasscode.value,
		})
	} catch {
		return
	}
	closeDisableForm()
}
</script>
