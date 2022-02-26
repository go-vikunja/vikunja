<template>
	<div>
		<message variant="success" text-align="center" class="mb-4" v-if="confirmedEmailSuccess">
			{{ $t('user.auth.confirmEmailSuccess') }}
		</message>
		<message variant="danger" v-if="errorMessage" class="mb-4">
			{{ errorMessage }}
		</message>
		<form @submit.prevent="submit" id="loginform" v-if="localAuthEnabled">
			<div class="field">
				<label class="label" for="username">{{ $t('user.auth.usernameEmail') }}</label>
				<div class="control">
					<input
						class="input" id="username"
						name="username"
						:placeholder="$t('user.auth.usernamePlaceholder')"
						ref="username"
						required
						type="text"
						autocomplete="username"
						v-focus
						@keyup.enter="submit"
						tabindex="1"
						@focusout="validateField('username')"
					/>
				</div>
				<p class="help is-danger" v-if="!usernameValid">
					{{ $t('user.auth.usernameRequired') }}
				</p>
			</div>
			<div class="field">
				<div class="label-with-link">
					<label class="label" for="password">{{ $t('user.auth.password') }}</label>
					<router-link
						:to="{ name: 'user.password-reset.request' }"
						class="reset-password-link"
						tabindex="6"
					>
						{{ $t('user.auth.forgotPassword') }}
					</router-link>
				</div>
				<password tabindex="2" @submit="submit" v-model="password" :validate-initially="validatePasswordInitially"/>
			</div>
			<div class="field" v-if="needsTotpPasscode">
				<label class="label" for="totpPasscode">{{ $t('user.auth.totpTitle') }}</label>
				<div class="control">
					<input
						autocomplete="one-time-code"
						class="input"
						id="totpPasscode"
						:placeholder="$t('user.auth.totpPlaceholder')"
						ref="totpPasscode"
						required
						type="text"
						v-focus
						@keyup.enter="submit"
						tabindex="3"
					/>
				</div>
			</div>
			<div class="field">
				<label class="label">
					<input type="checkbox" v-model="rememberMe" class="mr-1"/>
					{{ $t('user.auth.remember') }}
				</label>
			</div>

			<x-button
				@click="submit"
				:loading="loading"
				tabindex="4"
			>
				{{ $t('user.auth.login') }}
			</x-button>
			<p class="mt-2" v-if="registrationEnabled">
				{{ $t('user.auth.noAccountYet') }}
				<router-link
					:to="{ name: 'user.register' }"
					type="secondary"
					tabindex="5"
				>
					{{ $t('user.auth.createAccount') }}
				</router-link>
			</p>
		</form>

		<div
			v-if="hasOpenIdProviders"
			class="mt-4">
			<x-button
				@click="redirectToProvider(p)"
				v-for="(p, k) in openidConnect.providers"
				:key="k"
				variant="secondary"
				class="is-fullwidth mt-2"
			>
				{{ $t('user.auth.loginWith', {provider: p.name}) }}
			</x-button>
		</div>
	</div>
</template>

<script>
import {useDebounceFn} from '@vueuse/core'
import {mapState} from 'vuex'

import {HTTPFactory} from '@/http-common'
import {LOADING} from '@/store/mutation-types'
import {getErrorText} from '@/message'
import Message from '@/components/misc/message'
import {redirectToProvider} from '../../helpers/redirectToProvider'
import {getLastVisited, clearLastVisited} from '../../helpers/saveLastVisited'
import Password from '@/components/input/password'

export default {
	components: {
		Password,
		Message,
	},
	data() {
		return {
			confirmedEmailSuccess: false,
			errorMessage: '',
			usernameValid: true,
			password: '',
			validatePasswordInitially: false,
			rememberMe: false,
		}
	},
	beforeMount() {
		const HTTP = HTTPFactory()
		// Try to verify the email
		// FIXME: Why is this here? Can we find a better place for this?
		let emailVerifyToken = localStorage.getItem('emailConfirmToken')
		if (emailVerifyToken) {
			const cancel = this.setLoading()
			HTTP.post('user/confirm', {token: emailVerifyToken})
				.then(() => {
					localStorage.removeItem('emailConfirmToken')
					this.confirmedEmailSuccess = true
					cancel()
				})
				.catch(e => {
					cancel()
					this.errorMessage = e.response.data.message
				})
		}

		// Check if the user is already logged in, if so, redirect them to the homepage
		if (this.authenticated) {
			const last = getLastVisited()
			if (last !== null) {
				this.$router.push({
					name: last.name,
					params: last.params,
				})
				clearLastVisited()
			} else {
				this.$router.push({name: 'home'})
			}
		}
	},
	created() {
		this.setTitle(this.$t('user.auth.login'))
	},
	computed: {
		hasOpenIdProviders() {
			return this.openidConnect.enabled &&
				this.openidConnect.providers &&
				this.openidConnect.providers.length > 0
		},
		...mapState({
			registrationEnabled: state => state.config.registrationEnabled,
			loading: LOADING,
			needsTotpPasscode: state => state.auth.needsTotpPasscode,
			authenticated: state => state.auth.authenticated,
			localAuthEnabled: state => state.config.auth.local.enabled,
			openidConnect: state => state.config.auth.openidConnect,
		}),

		validateField() {
			// using computed so that debounced function definition stays
			return useDebounceFn((field) => {
				this[`${field}Valid`] = this.$refs[field]?.value !== ''
			}, 100)
		},
	},
	methods: {
		setLoading() {
			const timeout = setTimeout(() => {
				this.loading = true
			}, 100)
			return () => {
				clearTimeout(timeout)
				this.loading = false
			}
		},

		async submit() {
			this.errorMessage = ''
			// Some browsers prevent Vue bindings from working with autofilled values.
			// To work around this, we're manually getting the values here instead of relying on vue bindings.
			// For more info, see https://kolaente.dev/vikunja/frontend/issues/78
			const credentials = {
				username: this.$refs.username.value,
				password: this.password,
				longToken: this.rememberMe,
			}

			if (credentials.username === '' || credentials.password === '') {
				// Trigger the validation error messages
				this.validateField('username')
				this.validatePasswordInitially = true
				return
			}

			if (this.needsTotpPasscode) {
				credentials.totpPasscode = this.$refs.totpPasscode.value
			}

			try {
				await this.$store.dispatch('auth/login', credentials)
				this.$store.commit('auth/needsTotpPasscode', false)
			} catch (e) {
				if (e.response?.data.code === 1017 && !this.credentials.totpPasscode) {
					return
				}

				const err = getErrorText(e)
				this.errorMessage = typeof err[1] !== 'undefined' ? err[1] : err[0]
			}
		},

		redirectToProvider,
	},
}
</script>

<style lang="scss" scoped>
.button {
	margin: 0 0.4rem 0 0;
}

.reset-password-link {
	display: inline-block;
}

.label-with-link {
	display: flex;
	justify-content: space-between;
	margin-bottom: .5rem;

	.label {
		margin-bottom: 0;
	}
}
</style>
