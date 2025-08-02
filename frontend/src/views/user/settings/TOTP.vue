<template>
	<Card
		v-if="totpEnabled && isLocalUser"
		:title="$t('user.settings.totp.title')"
	>
		<XButton
			v-if="!totpEnrolled && totp.secret === ''"
			:loading="totpService.loading"
			@click="totpEnroll()"
		>
			{{ $t('user.settings.totp.enroll') }}
		</XButton>
		<template v-else-if="totp.secret !== '' && !totp.enabled">
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
			<div class="field">
				<label
					class="label"
					for="totpConfirmPasscode"
				>{{ $t('user.settings.totp.passcode') }}</label>
				<div class="control">
					<input
						id="totpConfirmPasscode"
						v-model="totpConfirmPasscode"
						autocomplete="one-time-code"
						class="input"
						:placeholder="$t('user.settings.totp.passcodePlaceholder')"
						type="text"
						inputmode="numeric"
						@keyup.enter="totpConfirm"
					>
				</div>
			</div>
			<XButton @click="totpConfirm">
				{{ $t('misc.confirm') }}
			</XButton>
		</template>
		<template v-else-if="totp.secret !== '' && totp.enabled">
			<p>
				{{ $t('user.settings.totp.setupSuccess') }}
			</p>
			<p v-if="!totpDisableForm">
				<XButton
					class="is-danger"
					@click="totpDisableForm = true"
				>
					{{ $t('misc.disable') }}
				</XButton>
			</p>
			<div v-if="totpDisableForm">
				<div class="field">
					<label
						class="label"
						for="currentPassword"
					>{{ $t('user.settings.totp.enterPassword') }}</label>
					<div class="control">
						<input
							id="currentPassword"
							v-model="totpDisablePassword"
							v-focus
							class="input"
							:placeholder="$t('user.settings.currentPasswordPlaceholder')"
							type="password"
							@keyup.enter="totpDisable"
						>
					</div>
				</div>
				<XButton
					class="is-danger"
					@click="totpDisable"
				>
					{{ $t('user.settings.totp.disable') }}
				</XButton>
				<XButton
					variant="tertiary"
					class="mis-2"
					@click="totpDisableForm = false"
				>
					{{ $t('misc.cancel') }}
				</XButton>
			</div>
		</template>
	</Card>
</template>


<script lang="ts" setup>
import {computed, ref, shallowReactive} from 'vue'
import {useI18n} from 'vue-i18n'

import TotpService from '@/services/totp'
import TotpModel from '@/models/totp'

import {success} from '@/message'

import {useTitle} from '@/composables/useTitle'
import {useConfigStore} from '@/stores/config'
import {useAuthStore} from '@/stores/auth'
import type {ITotp} from '@/modelTypes/ITotp'

defineOptions({name: 'UserSettingsTotp'})

const {t} = useI18n({useScope: 'global'})
useTitle(() => `${t('user.settings.totp.title')} - ${t('user.settings.title')}`)

const totpService = shallowReactive(new TotpService())
const totp = ref<ITotp>(new TotpModel())
const totpQR = ref('')
const totpEnrolled = ref(false)
const totpConfirmPasscode = ref('')
const totpDisableForm = ref(false)
const totpDisablePassword = ref('')

const configStore = useConfigStore()
const authStore = useAuthStore()
const totpEnabled = computed(() => configStore.totpEnabled)
const isLocalUser = computed(() => authStore.info?.isLocalUser)

totpStatus()

async function totpStatus() {
	if (!totpEnabled.value || !isLocalUser.value) {
		return
	}
	try {
		totp.value = await totpService.get({})
		totpSetQrCode()
	} catch(e: unknown) {
		// Error code 1016 means totp is not enabled, we don't need an error in that case.
		const err = e as {response?: {data?: {code?: number}}}
		if (err.response?.data?.code === 1016) {
			totpEnrolled.value = false
			return
		}

		throw e
	}
}

async function totpSetQrCode() {
	const qr = await totpService.qrcode()
	totpQR.value = window.URL.createObjectURL(qr)
}

async function totpEnroll() {
	totp.value = await totpService.enroll()
	totpEnrolled.value = true
	totpSetQrCode()
}

async function totpConfirm() {
	await totpService.enable({passcode: totpConfirmPasscode.value})
	totp.value.enabled = true
	success({message: t('user.settings.totp.confirmSuccess')})
}

async function totpDisable() {
	await totpService.disable({password: totpDisablePassword.value})
	totpEnrolled.value = false
	totp.value = new TotpModel()
	success({message: t('user.settings.totp.disableSuccess')})
}
</script>
