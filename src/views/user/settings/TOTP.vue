<template>
	<card :title="$t('user.settings.totp.title')" v-if="totpEnabled">
		<x-button
			:loading="totpService.loading"
			@click="totpEnroll()"
			v-if="!totpEnrolled && totp.secret === ''">
			{{ $t('user.settings.totp.enroll') }}
		</x-button>
		<template v-else-if="totp.secret !== '' && !totp.enabled">
			<p>
				{{ $t('user.settings.totp.finishSetupPart1') }}
				<strong>{{ totp.secret }}</strong><br/>
				{{ $t('user.settings.totp.finishSetupPart2') }}
			</p>
			<p>
				{{ $t('user.settings.totp.scanQR') }}<br/>
				<img :src="totpQR" alt=""/>
			</p>
			<div class="field">
				<label class="label" for="totpConfirmPasscode">{{ $t('user.settings.totp.passcode') }}</label>
				<div class="control">
					<input
						autocomplete="one-time-code"
						@keyup.enter="totpConfirm"
						class="input"
						id="totpConfirmPasscode"
						:placeholder="$t('user.settings.totp.passcodePlaceholder')"
						type="text"
						v-model="totpConfirmPasscode"/>
				</div>
			</div>
			<x-button @click="totpConfirm">{{ $t('misc.confirm') }}</x-button>
		</template>
		<template v-else-if="totp.secret !== '' && totp.enabled">
			<p>
				{{ $t('user.settings.totp.setupSuccess') }}
			</p>
			<p v-if="!totpDisableForm">
				<x-button @click="totpDisableForm = true" class="is-danger">{{ $t('misc.disable') }}</x-button>
			</p>
			<div v-if="totpDisableForm">
				<div class="field">
					<label class="label" for="currentPassword">{{ $t('user.settings.totp.enterPassword') }}</label>
					<div class="control">
						<input
							@keyup.enter="totpDisable"
							class="input"
							id="currentPassword"
							:placeholder="$t('user.settings.currentPasswordPlaceholder')"
							type="password"
							v-focus
							v-model="totpDisablePassword"/>
					</div>
				</div>
				<x-button @click="totpDisable" class="is-danger">
					{{ $t('user.settings.totp.disable') }}
				</x-button>
				<x-button @click="totpDisableForm = false" type="tertary" class="ml-2">
					{{ $t('misc.cancel') }}
				</x-button>
			</div>
		</template>
	</card>
</template>

<script>
import TotpService from '@/services/totp'
import TotpModel from '@/models/totp'
import {mapState} from 'vuex'

export default {
	name: 'user-settings-totp',
	data() {
		return {
			totpService: new TotpService(),
			totp: new TotpModel(),
			totpQR: '',
			totpEnrolled: false,
			totpConfirmPasscode: '',
			totpDisableForm: false,
			totpDisablePassword: '',
		}
	},
	created() {
		this.totpStatus()
	},
	computed: mapState({
		totpEnabled: state => state.config.totpEnabled,
	}),
	mounted() {
		this.setTitle(`${this.$t('user.settings.totp.title')} - ${this.$t('user.settings.title')}`)
	},
	methods: {
		async totpStatus() {
			if (!this.totpEnabled) {
				return
			}
			try {
				this.totp = await this.totpService.get()
				this.totpSetQrCode()
			} catch(e) {
				// Error code 1016 means totp is not enabled, we don't need an error in that case.
				if (e.response && e.response.data && e.response.data.code && e.response.data.code === 1016) {
					this.totpEnrolled = false
					return
				}

				throw e
			}
		},
		async totpSetQrCode() {
			const qr = await this.totpService.qrcode()
			this.totpQR = window.URL.createObjectURL(qr)
		},
		async totpEnroll() {
			this.totp = await this.totpService.enroll()
			this.totpEnrolled = true
			this.totpSetQrCode()
		},
		async totpConfirm() {
			await this.totpService.enable({passcode: this.totpConfirmPasscode})
			this.totp.enabled = true
			this.$message.success({message: this.$t('user.settings.totp.confirmSuccess')})
		},
		async totpDisable() {
			await this.totpService.disable({password: this.totpDisablePassword})
			this.totpEnrolled = false
			this.totp = new TotpModel()
			this.$message.success({message: this.$t('user.settings.totp.disableSuccess')})
		},
	},
}
</script>