<template>
	<div
		:class="{ 'is-loading': passwordUpdateService.loading || emailUpdateService.loading || totpService.loading }"
		class="loader-container is-max-width-desktop">
		<!-- General -->
		<card :title="$t('user.settings.general.title')" class="general-settings">
			<div class="field">
				<label class="label" for="newName">{{ $t('user.settings.general.name') }}</label>
				<div class="control">
					<input
						@keyup.enter="updateSettings"
						class="input"
						id="newName"
						:placeholder="$t('user.settings.general.newName')"
						type="text"
						v-model="settings.name"/>
				</div>
			</div>
			<div class="field">
				<label class="label">
					{{ $t('user.settings.general.defaultList') }}
				</label>
				<list-search v-model="defaultList"/>
			</div>
			<div class="field">
				<label class="checkbox">
					<input type="checkbox" v-model="settings.emailRemindersEnabled"/>
					{{ $t('user.settings.general.emailReminders') }}
				</label>
			</div>
			<div class="field">
				<label class="checkbox">
					<input type="checkbox" v-model="settings.overdueTasksRemindersEnabled"/>
					{{ $t('user.settings.general.overdueReminders') }}
				</label>
			</div>
			<div class="field">
				<label class="checkbox">
					<input type="checkbox" v-model="settings.discoverableByName"/>
					{{ $t('user.settings.general.discoverableByName') }}
				</label>
			</div>
			<div class="field">
				<label class="checkbox">
					<input type="checkbox" v-model="settings.discoverableByEmail"/>
					{{ $t('user.settings.general.discoverableByEmail') }}
				</label>
			</div>
			<div class="field">
				<label class="checkbox">
					<input type="checkbox" v-model="playSoundWhenDone"/>
					{{ $t('user.settings.general.playSoundWhenDone') }}
				</label>
			</div>
			<div class="field">
				<label class="is-flex is-align-items-center">
					<span>
						{{ $t('user.settings.general.weekStart') }}
					</span>
					<div class="select ml-2">
						<select v-model.number="settings.weekStart">
							<option value="0">{{ $t('user.settings.general.weekStartSunday') }}</option>
							<option value="1">{{ $t('user.settings.general.weekStartMonday') }}</option>
						</select>
					</div>
				</label>
			</div>
			<div class="field">
				<label class="is-flex is-align-items-center">
					<span>
						{{ $t('user.settings.general.language') }}
					</span>
					<div class="select ml-2">
						<select v-model="language">
							<option :value="lang.code" v-for="lang in availableLanguages" :key="lang.code">{{ lang.title }}</option>
						</select>
					</div>
				</label>
			</div>

			<x-button
				:loading="userSettingsService.loading"
				@click="updateSettings()"
				class="is-fullwidth mt-4"
			>
				{{ $t('misc.save') }}
			</x-button>
		</card>

		<!-- Avatar -->
		<avatar-settings/>

		<!-- Password update -->
		<card :title="$t('user.settings.newPasswordTitle')">
			<form @submit.prevent="updatePassword()">
				<div class="field">
					<label class="label" for="newPassword">{{ $t('user.settings.newPassword') }}</label>
					<div class="control">
						<input
							@keyup.enter="updatePassword"
							class="input"
							id="newPassword"
							:placeholder="$t('user.auth.passwordPlaceholder')"
							type="password"
							v-model="passwordUpdate.newPassword"/>
					</div>
				</div>
				<div class="field">
					<label class="label" for="newPasswordConfirm">{{ $t('user.settings.newPasswordConfirm') }}</label>
					<div class="control">
						<input
							@keyup.enter="updatePassword"
							class="input"
							id="newPasswordConfirm"
							:placeholder="$t('user.auth.passwordPlaceholder')"
							type="password"
							v-model="passwordConfirm"/>
					</div>
				</div>
				<div class="field">
					<label class="label" for="currentPassword">{{ $t('user.settings.currentPassword') }}</label>
					<div class="control">
						<input
							@keyup.enter="updatePassword"
							class="input"
							id="currentPassword"
							:placeholder="$t('user.settings.currentPasswordPlaceholder')"
							type="password"
							v-model="passwordUpdate.oldPassword"/>
					</div>
				</div>
			</form>

			<x-button
				:loading="passwordUpdateService.loading"
				@click="updatePassword()"
				class="is-fullwidth mt-4">
				{{ $t('misc.save') }}
			</x-button>
		</card>

		<!-- Update E-Mail -->
		<card :title="$t('user.settings.updateEmailTitle')">
			<form @submit.prevent="updateEmail()">
				<div class="field">
					<label class="label" for="newEmail">{{ $t('user.settings.updateEmailNew') }}</label>
					<div class="control">
						<input
							@keyup.enter="updateEmail"
							class="input"
							id="newEmail"
							:placeholder="$t('user.auth.emailPlaceholder')"
							type="email"
							v-model="emailUpdate.newEmail"/>
					</div>
				</div>
				<div class="field">
					<label class="label" for="currentPasswordEmail">{{ $t('user.settings.currentPassword') }}</label>
					<div class="control">
						<input
							@keyup.enter="updateEmail"
							class="input"
							id="currentPasswordEmail"
							:placeholder="$t('user.settings.currentPasswordPlaceholder')"
							type="password"
							v-model="emailUpdate.password"/>
					</div>
				</div>
			</form>

			<x-button
				:loading="emailUpdateService.loading"
				@click="updateEmail()"
				class="is-fullwidth mt-4">
				{{ $t('misc.save') }}
			</x-button>
		</card>

		<!-- TOTP -->
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
							@keyup.enter="totpConfirm()"
							class="input"
							id="totpConfirmPasscode"
							:placeholder="$t('user.settings.totp.passcodePlaceholder')"
							type="text"
							v-model="totpConfirmPasscode"/>
					</div>
				</div>
				<x-button @click="totpConfirm()">{{ $t('misc.confirm') }}</x-button>
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
					<x-button @click="totpDisable()" class="is-danger">
						{{ $t('user.settings.totp.disable') }}
					</x-button>
				</div>
			</template>
		</card>
		
		<!-- Data export -->
		<data-export/>

		<!-- Migration -->
		<card :title="$t('migrate.title')" v-if="migratorsEnabled">
			<x-button
				:to="{name: 'migrate.start'}"
			>
				{{ $t('migrate.import') }}
			</x-button>
		</card>

		<!-- Account deletion -->
		<user-settings-deletion id="deletion"/>

		<!-- Caldav -->
		<card v-if="caldavEnabled" :title="$t('user.settings.caldav.title')">
			<p>
				{{ $t('user.settings.caldav.howTo') }}
			</p>
			<div class="field has-addons no-input-mobile">
				<div class="control is-expanded">
					<input type="text" v-model="caldavUrl" class="input" readonly/>
				</div>
				<div class="control">
					<x-button
						@click="copy(caldavUrl)"
						:shadow="false"
						v-tooltip="$t('misc.copy')"
						icon="paste"
					/>
				</div>
			</div>
			<p>
				<a href="https://vikunja.io/docs/caldav/" rel="noreferrer noopener nofollow" target="_blank">
					{{ $t('user.settings.caldav.more') }}
				</a>
			</p>
		</card>
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
import {playSoundWhenDoneKey} from '@/helpers/playPop'
import {availableLanguages, saveLanguage, getCurrentLanguage} from '../../i18n/setup'

import {mapState} from 'vuex'

import AvatarSettings from '../../components/user/avatar-settings.vue'
import copy from 'copy-to-clipboard'
import ListSearch from '@/components/tasks/partials/listSearch.vue'
import UserSettingsDeletion from '../../components/user/settings/deletion'
import DataExport from '../../components/user/settings/data-export'

export default {
	name: 'Settings',
	data() {
		return {
			passwordUpdateService: new PasswordUpdateService(),
			passwordUpdate: new PasswordUpdateModel(),
			passwordConfirm: '',

			emailUpdateService: new EmailUpdateService(),
			emailUpdate: new EmailUpdateModel(),

			totpService: new TotpService(),
			totp: new TotpModel(),
			totpQR: '',
			totpEnrolled: false,
			totpConfirmPasscode: '',
			totpDisableForm: false,
			totpDisablePassword: '',
			playSoundWhenDone: false,
			language: getCurrentLanguage(),

			settings: UserSettingsModel,
			userSettingsService: new UserSettingsService(),

			defaultList: null,
		}
	},
	components: {
		UserSettingsDeletion,
		ListSearch,
		AvatarSettings,
		DataExport,
	},
	created() {
		this.settings = this.$store.state.auth.settings

		this.playSoundWhenDone = localStorage.getItem(playSoundWhenDoneKey) === 'true' || localStorage.getItem(playSoundWhenDoneKey) === null

		this.defaultList = this.$store.getters['lists/getListById'](this.settings.defaultListId)

		this.totpStatus()
	},
	mounted() {
		this.setTitle(this.$t('user.settings.title'))
		this.anchorHashCheck()
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
		availableLanguages() {
			return Object.entries(availableLanguages)
				.map(l => ({code: l[0], title: l[1]}))
				.sort((a, b) => a.title > b.title)
		},
		...mapState({
			totpEnabled: state => state.config.totpEnabled,
			migratorsEnabled: state => state.config.availableMigrators !== null && state.config.availableMigrators.length > 0,
			caldavEnabled: state => state.config.caldavEnabled,
			userInfo: state => state.auth.info,
		}),
	},
	methods: {
		copy,

		updatePassword() {
			if (this.passwordConfirm !== this.passwordUpdate.newPassword) {
				this.$message.error({message: this.$t('user.settings.passwordsDontMatch')})
				return
			}

			this.passwordUpdateService.update(this.passwordUpdate)
				.then(() => {
					this.$message.success({message: this.$t('user.settings.passwordUpdateSuccess')})
				})
				.catch(e => this.$message.error(e))
		},
		updateEmail() {
			this.emailUpdateService.update(this.emailUpdate)
				.then(() => {
					this.$message.success({message: this.$t('user.settings.updateEmailSuccess')})
				})
				.catch(e => this.$message.error(e))
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

					this.$message.error(e)
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
				.catch(e => this.$message.error(e))
		},
		totpConfirm() {
			this.totpService.enable({passcode: this.totpConfirmPasscode})
				.then(() => {
					this.$set(this.totp, 'enabled', true)
					this.$message.success({message: this.$t('user.settings.totp.confirmSuccess')})
				})
				.catch(e => this.$message.error(e))
		},
		totpDisable() {
			this.totpService.disable({password: this.totpDisablePassword})
				.then(() => {
					this.totpEnrolled = false
					this.$set(this, 'totp', new TotpModel())
					this.$message.success({message: this.$t('user.settings.totp.disableSuccess')})
				})
				.catch(e => this.$message.error(e))
		},
		updateSettings() {
			localStorage.setItem(playSoundWhenDoneKey, this.playSoundWhenDone)
			saveLanguage(this.language)
			this.settings.defaultListId = this.defaultList ? this.defaultList.id : 0

			this.userSettingsService.update(this.settings)
				.then(() => {
					this.$store.commit('auth/setUserSettings', this.settings)
					this.$message.success({message: this.$t('user.settings.general.savedSuccess')})
				})
				.catch(e => this.$message.error(e))
		},
		anchorHashCheck() {
			if (window.location.hash === this.$route.hash) {
				const el = document.getElementById(this.$route.hash.slice(1))
				if (el) {
					window.scrollTo(0, el.offsetTop)
				}
			}
		},
	},
}
</script>
