<template>
	<Card
		v-if="configStore.totp_enabled && isLocalUser"
		:title="$t('user.settings.totp.title')"
	>
		<XButton
			v-if="!totp.enabled && !totp.secret"
			:loading="enrollMutation.isPending.value"
			@click="enrollMutation.mutate()"
		>
			{{ $t('user.settings.totp.enroll') }}
		</XButton>
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
			<FormField
				id="totpConfirmPasscode"
				v-model="totpConfirmPasscode"
				:label="$t('user.settings.totp.passcode')"
				autocomplete="one-time-code"
				:placeholder="$t('user.settings.totp.passcodePlaceholder')"
				type="text"
				inputmode="numeric"
				@keyup.enter="totpConfirm"
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
				<FormField
					id="currentPassword"
					v-model="totpDisablePassword"
					v-focus
					:label="$t('user.settings.totp.enterPassword')"
					:placeholder="$t('user.settings.currentPasswordPlaceholder')"
					type="password"
					@keyup.enter="totpDisable"
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
import {totpQuery, totpQrQuery, useEnrollTotpMutation, useEnableTotpMutation, useDisableTotpMutation} from '@/client/queries/totp'
import FormField from '@/components/input/FormField.vue'


import {useTitle} from '@/composables/useTitle'
import {useConfigStore} from '@/stores/config'
import {useAuthStore} from '@/stores/auth'

defineOptions({name: 'UserSettingsTotp'})

const {t} = useI18n({useScope: 'global'})
useTitle(() => `${t('user.settings.totp.title')} - ${t('user.settings.title')}`)

const totpConfirmPasscode = ref('')
const totpDisableForm = ref(false)
const totpDisablePassword = ref('')

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

async function totpConfirm() {
	if (enableMutation.isPending.value) return
	try {
		await enableMutation.mutateAsync(totpConfirmPasscode.value)
		await authStore.logout()
	} catch { return }
}

function closeDisableForm() {
	totpDisableForm.value = false
	totpDisablePassword.value = ''
}

async function totpDisable() {
	if (disableMutation.isPending.value) return
	try {
		await disableMutation.mutateAsync(totpDisablePassword.value)
	} catch {
		return
	} finally {
		// Evicts the plaintext password from the mutation cache.
		disableMutation.reset()
	}
	closeDisableForm()
}
</script>
