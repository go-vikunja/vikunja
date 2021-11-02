<template>
	<div>
		<h2 class="title has-text-centered">{{ $t('user.auth.register') }}</h2>
		<div class="box">
			<form @submit.prevent="submit" id="registerform">
				<div class="field">
					<label class="label" for="username">{{ $t('user.auth.username') }}</label>
					<div class="control">
						<input
							class="input"
							id="username"
							name="username"
							:placeholder="$t('user.auth.usernamePlaceholder')"
							required
							type="text"
							autocomplete="username"
							v-focus
							v-model="credentials.username"
							@keyup.enter="submit"
						/>
					</div>
				</div>
				<div class="field">
					<label class="label" for="email">{{ $t('user.auth.email') }}</label>
					<div class="control">
						<input
							class="input"
							id="email"
							name="email"
							:placeholder="$t('user.auth.emailPlaceholder')"
							required
							type="email"
							v-model="credentials.email"
							@keyup.enter="submit"
						/>
					</div>
				</div>
				<div class="field">
					<label class="label" for="password1">{{ $t('user.auth.password') }}</label>
					<div class="control">
						<input
							class="input"
							id="password1"
							name="password1"
							:placeholder="$t('user.auth.passwordPlaceholder')"
							required
							type="password"
							autocomplete="new-password"
							v-model="credentials.password"
							@keyup.enter="submit"
						/>
					</div>
				</div>
				<div class="field">
					<label class="label" for="password2">{{ $t('user.auth.passwordRepeat') }}</label>
					<div class="control">
						<input
							class="input"
							id="password2"
							name="password2"
							:placeholder="$t('user.auth.passwordPlaceholder')"
							required
							type="password"
							autocomplete="new-password"
							v-model="credentials.password2"
							@keyup.enter="submit"
						/>
					</div>
				</div>

				<div class="field is-grouped">
					<div class="control">
						<x-button
							:loading="loading"
							id="register-submit"
							@click="submit"
							class="mr-2"
						>
							{{ $t('user.auth.register') }}
						</x-button>
						<x-button :to="{ name: 'user.login' }" type="secondary">
							{{ $t('user.auth.login') }}
						</x-button>
					</div>
				</div>
				<div class="notification is-info" v-if="loading">
					{{ $t('misc.loading') }}
				</div>
				<div class="notification is-danger" v-if="errorMessage !== ''">
					{{ errorMessage }}
				</div>
			</form>
			<legal/>
		</div>
	</div>
</template>

<script>
import router from '../../router'
import {mapState} from 'vuex'
import {LOADING} from '@/store/mutation-types'
import Legal from '../../components/misc/legal'

export default {
	components: {
		Legal,
	},
	data() {
		return {
			credentials: {
				username: '',
				email: '',
				password: '',
				password2: '',
			},
			errorMessage: '',
		}
	},
	beforeMount() {
		// Check if the user is already logged in, if so, redirect them to the homepage
		if (this.authenticated) {
			router.push({name: 'home'})
		}
	},
	mounted() {
		this.setTitle(this.$t('user.auth.register'))
	},
	computed: mapState({
		authenticated: state => state.auth.authenticated,
		loading: LOADING,
	}),
	methods: {
		async submit() {
			this.errorMessage = ''

			if (this.credentials.password2 !== this.credentials.password) {
				this.errorMessage = this.$t('user.auth.passwordsDontMatch')
				return
			}

			const credentials = {
				username: this.credentials.username,
				email: this.credentials.email,
				password: this.credentials.password,
			}

			try {
				await this.$store.dispatch('auth/register', credentials)
			} catch(e) {
				this.errorMessage = e.message
			}
		},
	},
}
</script>
