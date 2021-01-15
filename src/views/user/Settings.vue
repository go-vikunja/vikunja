<template>
	<div
		:class="{ 'is-loading': passwordUpdateService.loading || emailUpdateService.loading || totpService.loading }"
		class="loader-container is-max-width-desktop">
		<!-- Password update -->
		<div class="card">
			<header class="card-header">
				<p class="card-header-title">
					Update Your Password
				</p>
			</header>
			<div class="card-content">
				<div class="content">
					<form @submit.prevent="updatePassword()">
						<div class="field">
							<label class="label" for="newPassword">New Password</label>
							<div class="control">
								<input
									@keyup.enter="updatePassword"
									class="input"
									id="newPassword"
									placeholder="The new password..."
									type="password"
									v-model="passwordUpdate.newPassword"/>
							</div>
						</div>
						<div class="field">
							<label class="label" for="newPasswordConfirm">New Password Confirmation</label>
							<div class="control">
								<input
									@keyup.enter="updatePassword"
									class="input"
									id="newPasswordConfirm"
									placeholder="Confirm your new password..."
									type="password"
									v-model="passwordConfirm"/>
							</div>
						</div>
						<div class="field">
							<label class="label" for="currentPassword">Current Password</label>
							<div class="control">
								<input
									@keyup.enter="updatePassword"
									class="input"
									id="currentPassword"
									placeholder="Your current password"
									type="password"
									v-model="passwordUpdate.oldPassword"/>
							</div>
						</div>
					</form>

					<div class="bigbuttons">
						<button :class="{ 'is-loading': passwordUpdateService.loading}" @click="updatePassword()"
								class="button is-primary is-fullwidth">
							Save
						</button>
					</div>
				</div>
			</div>
		</div>

		<!-- Update E-Mail -->
		<div class="card">
			<header class="card-header">
				<p class="card-header-title">
					Update Your E-Mail Address
				</p>
			</header>
			<div class="card-content">
				<div class="content">
					<form @submit.prevent="updateEmail()">
						<div class="field">
							<label class="label" for="newEmail">New Email Address</label>
							<div class="control">
								<input
									@keyup.enter="updateEmail"
									class="input"
									id="newEmail"
									placeholder="The new email address..."
									type="email"
									v-model="emailUpdate.newEmail"/>
							</div>
						</div>
						<div class="field">
							<label class="label" for="currentPassword">Current Password</label>
							<div class="control">
								<input
									@keyup.enter="updateEmail"
									class="input"
									id="currentPassword"
									placeholder="Your current password"
									type="password"
									v-model="emailUpdate.password"/>
							</div>
						</div>
					</form>

					<div class="bigbuttons">
						<button :class="{ 'is-loading': emailUpdateService.loading}" @click="updateEmail()"
								class="button is-primary is-fullwidth">
							Save
						</button>
					</div>
				</div>
			</div>
		</div>

		<!-- General -->
		<div class="card update-name">
			<header class="card-header">
				<p class="card-header-title">
					General Settings
				</p>
			</header>
			<div class="card-content">
				<div class="content">
					<div class="field">
						<label class="label" for="newName">Name</label>
						<div class="control">
							<input
								@keyup.enter="updateSettings"
								class="input"
								id="newName"
								placeholder="The new name"
								type="text"
								v-model="settings.name"/>
						</div>
					</div>
					<div class="field">
						<label class="checkbox">
							<input type="checkbox" v-model="settings.emailRemindersEnabled"/>
							Send me Reminders for tasks via Email
						</label>
					</div>

					<div class="bigbuttons">
						<button :class="{ 'is-loading': userSettingsService.loading}" @click="updateSettings()"
								class="button is-primary is-fullwidth">
							Save
						</button>
					</div>
				</div>
			</div>
		</div>

		<!-- Avatar -->
		<avatar-settings/>

		<!-- TOTP -->
		<div class="card" v-if="totpEnabled">
			<header class="card-header">
				<p class="card-header-title">
					Two Factor Authentication
				</p>
			</header>
			<div class="card-content">
				<a
					:class="{ 'is-loading': totpService.loading }"
					@click="totpEnroll()"
					class="button is-primary"
					v-if="!totpEnrolled && totp.secret === ''">
					Enroll
				</a>
				<div class="content" v-else-if="totp.secret !== '' && !totp.enabled">
					<p>
						To finish your setup, use this secret in your totp app (Google Authenticator or similar):
						<strong>{{ totp.secret }}</strong><br/>
						After that, enter a code from your app below.
					</p>
					<p>
						Alternatively you can scan this QR code:<br/>
						<img :src="totpQR" alt=""/>
					</p>
					<div class="field">
						<label class="label" for="totpConfirmPasscode">Passcode</label>
						<div class="control">
							<input
								@keyup.enter="totpConfirm()"
								class="input"
								id="totpConfirmPasscode"
								placeholder="A code generated by your totp application"
								type="text"
								v-model="totpConfirmPasscode"/>
						</div>
					</div>
					<a @click="totpConfirm()" class="button is-primary">Confirm</a>
				</div>
				<div class="content" v-else-if="totp.secret !== '' && totp.enabled">
					<p>
						You've sucessfully set up two factor authentication!
					</p>
					<p v-if="!totpDisableForm">
						<a @click="totpDisableForm = true" class="button is-danger">Disable</a>
					</p>
					<div v-if="totpDisableForm">
						<div class="field">
							<label class="label" for="currentPassword">Please Enter Your Password</label>
							<div class="control">
								<input
									@keyup.enter="totpDisable"
									class="input"
									id="currentPassword"
									placeholder="Your current password"
									type="password"
									v-focus
									v-model="totpDisablePassword"/>
							</div>
						</div>
						<a @click="totpDisable()" class="button is-danger">Disable two factor authentication</a>
					</div>
				</div>
			</div>
		</div>

		<!-- Migration -->
		<div class="card" v-if="migratorsEnabled">
			<header class="card-header">
				<p class="card-header-title">
					Migrate from other services to Vikunja
				</p>
			</header>
			<div class="card-content">
				<router-link
					:to="{name: 'migrate.start'}"
					class="button is-primary"
					v-if="migratorsEnabled"
				>
					Import your data into Vikunja
				</router-link>
			</div>
		</div>

		<!-- Caldav -->
		<div class="card" v-if="caldavEnabled">
			<header class="card-header">
				<p class="card-header-title">
					Caldav
				</p>
			</header>
			<div class="card-content content">
				<p>
					You can connect Vikunja to caldav clients to view and manage all tasks from different clients.
					Enter this url into your client:
				</p>
				<div class="field has-addons no-input-mobile">
					<div class="control is-expanded">
						<input type="text" v-model="caldavUrl" class="input" readonly/>
					</div>
					<div class="control">
						<a @click="copy(caldavUrl)" class="button is-success noshadow" v-tooltip="'Copy to clipboard'">
							<span class="icon">
								<icon icon="paste"/>
							</span>
						</a>
					</div>
				</div>
				<p>
					<a href="https://vikunja.io/docs/caldav/" target="_blank">
						More information about caldav in Vikunja
					</a>
				</p>
			</div>
		</div>
	</div>
</template>

<script>
import PasswordUpdateModel from '../../models/passwordUpdate'
import PasswordUpdateService from '../../services/passwordUpdateService'
import EmailUpdateService from '../../services/emailUpdate'
import EmailUpdateModel from '../../models/emailUpdate'
import TotpModel from '../../models/totp'
import TotpService from '../../services/totp'
import UserSettingsService from '../../services/userSettings'
import UserSettingsModel from '../../models/userSettings'

import {mapState} from 'vuex'

import AvatarSettings from '../../components/user/avatar-settings'
import copy from 'copy-to-clipboard'

export default {
	name: 'Settings',
	data() {
		return {
			passwordUpdateService: PasswordUpdateService,
			passwordUpdate: PasswordUpdateModel,
			passwordConfirm: '',

			emailUpdateService: EmailUpdateService,
			emailUpdate: EmailUpdateModel,

			totpService: TotpService,
			totp: TotpModel,
			totpQR: '',
			totpEnrolled: false,
			totpConfirmPasscode: '',
			totpDisableForm: false,
			totpDisablePassword: '',

			settings: UserSettingsModel,
			userSettingsService: UserSettingsService,
		}
	},
	components: {
		AvatarSettings,
	},
	created() {
		this.passwordUpdateService = new PasswordUpdateService()
		this.passwordUpdate = new PasswordUpdateModel()

		this.emailUpdateService = new EmailUpdateService()
		this.emailUpdate = new EmailUpdateModel()

		this.totpService = new TotpService()
		this.totp = new TotpModel()

		this.userSettingsService = new UserSettingsService()
		this.settings = new UserSettingsModel({
			name: this.$store.state.auth.info.name,
			emailRemindersEnabled: this.$store.state.auth.info.emailRemindersEnabled ?? false,
		})

		this.totpStatus()
	},
	mounted() {
		this.setTitle('Settings')
	},
	computed: {
		caldavUrl() {
			let apiBase = window.API_URL.replace('/api/v1', '')
			if (apiBase === '') { // Frontend and api on the same host which means we need to prefix the frontend url
				apiBase = this.$store.state.config.frontendUrl
			}
			if (apiBase.endsWith('/')) {
				apiBase = apiBase.substr(0, apiBase.length - 1)
			}

			return `${apiBase}/dav/principals/${this.userInfo.username}/`
		},
		...mapState({
			totpEnabled: state => state.config.totpEnabled,
			migratorsEnabled: state => state.config.availableMigrators !== null && state.config.availableMigrators.length > 0,
			caldavEnabled: state => state.config.caldavEnabled,
			userInfo: state => state.auth.info,
		})
	},
	methods: {
		updatePassword() {
			if (this.passwordConfirm !== this.passwordUpdate.newPassword) {
				this.error({message: 'The new password and its confirmation don\'t match.'}, this)
				return
			}

			this.passwordUpdateService.update(this.passwordUpdate)
				.then(() => {
					this.success({message: 'The password was successfully updated.'}, this)
				})
				.catch(e => this.error(e, this))
		},
		updateEmail() {
			this.emailUpdateService.update(this.emailUpdate)
				.then(() => {
					this.success({message: 'Your email address was successfully updated. We\'ve sent you a link to confirm it.'}, this)
				})
				.catch(e => this.error(e, this))
		},
		totpStatus() {
			if (!this.totpEnabled) {
				return
			}
			this.totpService.get()
				.then(r => {
					this.$set(this, 'totp', r)
					this.totpSetQrCode()
				})
				.catch(e => {
					// Error code 1016 means totp is not enabled, we don't need an error in that case.
					if (e.response && e.response.data && e.response.data.code && e.response.data.code === 1016) {
						this.totpEnrolled = false
						return
					}

					this.error(e, this)
				})
		},
		totpSetQrCode() {
			this.totpService.qrcode()
				.then(qr => {
					const urlCreator = window.URL || window.webkitURL
					this.totpQR = urlCreator.createObjectURL(qr)
				})
		},
		totpEnroll() {
			this.totpService.enroll()
				.then(r => {
					this.totpEnrolled = true
					this.$set(this, 'totp', r)
					this.totpSetQrCode()
				})
				.catch(e => this.error(e, this))
		},
		totpConfirm() {
			this.totpService.enable({passcode: this.totpConfirmPasscode})
				.then(() => {
					this.$set(this.totp, 'enabled', true)
					this.success({message: 'You\'ve successfully confirmed your totp setup and can use it from now on!'}, this)
				})
				.catch(e => this.error(e, this))
		},
		totpDisable() {
			this.totpService.disable({password: this.totpDisablePassword})
				.then(() => {
					this.totpEnrolled = false
					this.$set(this, 'totp', new TotpModel())
					this.success({message: 'Two factor authentication was sucessfully disabled.'}, this)
				})
				.catch(e => this.error(e, this))
		},
		updateSettings() {
			this.userSettingsService.update(this.settings)
				.then(() => {
					this.$store.commit('auth/setUserSettings', this.settings)
					this.success({message: 'The name was successfully changed.'}, this)
				})
				.catch(e => this.error(e, this))
		},
		copy(text) {
			copy(text)
		},
	},
}
</script>
