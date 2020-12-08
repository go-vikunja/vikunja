<template>
	<div>
		<h2 class="title has-text-centered">Login</h2>
		<div class="box">
			<div class="notification is-success has-text-centered" v-if="confirmedEmailSuccess">
				You successfully confirmed your email! You can log in now.
			</div>
			<api-config/>
			<form @submit.prevent="submit" id="loginform" v-if="localAuthEnabled">
				<div class="field">
					<label class="label" for="username">Username</label>
					<div class="control">
						<input
							class="input" id="username"
							name="username"
							placeholder="e.g. frederick"
							ref="username"
							required
							type="text"
							autocomplete="username"
							v-focus
						/>
					</div>
				</div>
				<div class="field">
					<label class="label" for="password">Password</label>
					<div class="control">
						<input
							class="input"
							id="password"
							name="password"
							placeholder="e.g. ••••••••••••"
							ref="password"
							required
							type="password"
							autocomplete="current-password"
						/>
					</div>
				</div>
				<div class="field" v-if="needsTotpPasscode">
					<label class="label" for="totpPasscode">Two Factor Authentication Code</label>
					<div class="control">
						<input
							class="input"
							id="totpPasscode"
							placeholder="e.g. 123456"
							ref="totpPasscode"
							required
							type="text"
							v-focus
						/>
					</div>
				</div>

				<div class="field is-grouped login-buttons">
					<div class="control is-expanded">
						<button class="button is-primary" type="submit" v-bind:class="{ 'is-loading': loading}">
							Login
						</button>
						<router-link :to="{ name: 'user.register' }" class="button" v-if="registrationEnabled">
							Register
						</router-link>
					</div>
					<div class="control">
						<router-link :to="{ name: 'user.password-reset.request' }" class="reset-password-link">
							Reset your password
						</router-link>
					</div>
				</div>
				<div class="notification is-danger" v-if="errorMessage">
					{{ errorMessage }}
				</div>
			</form>

			<div v-if="openidConnect.enabled && openidConnect.providers.length > 0" class="mt-4">
				<a @click="redirectToProvider(p)" v-for="(p, k) in openidConnect.providers" :key="k" class="button is-fullwidth">
					Log in with {{ p.name }}
				</a>
			</div>

			<legal/>
		</div>
	</div>
</template>

<script>
import {mapState} from 'vuex'

import router from '../../router'
import {HTTPFactory} from '@/http-common'
import message from '../../message'
import {ERROR_MESSAGE, LOADING} from '@/store/mutation-types'
import legal from '../../components/misc/legal'
import ApiConfig from '@/components/misc/api-config'

export default {
	components: {
		ApiConfig,
		legal,
	},
	data() {
		return {
			confirmedEmailSuccess: false,
		}
	},
	beforeMount() {
		const HTTP = HTTPFactory()
		// Try to verify the email
		// FIXME: Why is this here? Can we find a better place for this?
		let emailVerifyToken = localStorage.getItem('emailConfirmToken')
		if (emailVerifyToken) {
			const cancel = message.setLoading(this)
			HTTP.post(`user/confirm`, {token: emailVerifyToken})
				.then(() => {
					localStorage.removeItem('emailConfirmToken')
					this.confirmedEmailSuccess = true
					cancel()
				})
				.catch(e => {
					cancel()
					this.$store.commit(ERROR_MESSAGE, e.response.data.message)
				})
		}

		// Check if the user is already logged in, if so, redirect him to the homepage
		if (this.authenticated) {
			router.push({name: 'home'})
		}
	},
	created() {
		this.setTitle('Login')
	},
	computed: mapState({
		registrationEnabled: state => state.config.registrationEnabled,
		loading: LOADING,
		errorMessage: ERROR_MESSAGE,
		needsTotpPasscode: state => state.auth.needsTotpPasscode,
		authenticated: state => state.auth.authenticated,
		localAuthEnabled: state => state.config.auth.local.enabled,
		openidConnect: state => state.config.auth.openidConnect,
	}),
	methods: {
		submit() {
			this.$store.commit(ERROR_MESSAGE, '')
			// Some browsers prevent Vue bindings from working with autofilled values.
			// To work around this, we're manually getting the values here instead of relying on vue bindings.
			// For more info, see https://kolaente.dev/vikunja/frontend/issues/78
			const credentials = {
				username: this.$refs.username.value,
				password: this.$refs.password.value,
			}

			if (this.needsTotpPasscode) {
				credentials.totpPasscode = this.$refs.totpPasscode.value
			}

			this.$store.dispatch('auth/login', credentials)
				.then(() => {
					router.push({name: 'home'})
				})
				.catch(() => {
				})
		},
		redirectToProvider(provider) {
			const state = Math.random().toString(36).substring(2, 24)
			localStorage.setItem('state', state)

			window.location.href = `${provider.authUrl}?client_id=${provider.clientId}&redirect_uri=${this.openidConnect.redirectUrl}${provider.key}&response_type=code&scope=openid&state=${state}`
		},
	},
}
</script>

<style scoped>
.button {
	margin: 0 0.4em 0 0;
}

.reset-password-link {
	display: inline-block;
	padding-top: 5px;
}
</style>
