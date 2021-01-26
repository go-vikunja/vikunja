<template>
	<div>
		<h2 class="title has-text-centered">Register</h2>
		<div class="box">
			<form @submit.prevent="submit" id="registerform">
				<div class="field">
					<label class="label" for="username">Username</label>
					<div class="control">
						<input
							class="input"
							id="username"
							name="username"
							placeholder="e.g. frederick"
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
					<label class="label" for="email">E-mail address</label>
					<div class="control">
						<input
							class="input"
							id="email"
							name="email"
							placeholder="e.g. frederic@vikunja.io"
							required
							type="email"
							v-model="credentials.email"
							@keyup.enter="submit"
						/>
					</div>
				</div>
				<div class="field">
					<label class="label" for="password1">Password</label>
					<div class="control">
						<input
							class="input"
							id="password1"
							name="password1"
							placeholder="e.g. ••••••••••••"
							required
							type="password"
							autocomplete="new-password"
							v-model="credentials.password"
							@keyup.enter="submit"
						/>
					</div>
				</div>
				<div class="field">
					<label class="label" for="password2">Retype your password</label>
					<div class="control">
						<input
							class="input"
							id="password2"
							name="password2"
							placeholder="e.g. ••••••••••••"
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
							Register
						</x-button>
						<x-button :to="{ name: 'user.login' }" type="secondary">Login</x-button>
					</div>
				</div>
				<div class="notification is-info" v-if="loading">
					Loading...
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
import {ERROR_MESSAGE, LOADING} from '@/store/mutation-types'
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
		}
	},
	beforeMount() {
		// Check if the user is already logged in, if so, redirect him to the homepage
		if (this.authenticated) {
			router.push({name: 'home'})
		}
	},
	mounted() {
		this.setTitle('Register')
	},
	computed: mapState({
		authenticated: state => state.auth.authenticated,
		loading: LOADING,
		errorMessage: ERROR_MESSAGE,
	}),
	methods: {
		submit() {
			this.$store.commit(LOADING, true)
			this.$store.commit(ERROR_MESSAGE, '')

			if (this.credentials.password2 !== this.credentials.password) {
				this.$store.commit(ERROR_MESSAGE, 'Passwords don\'t match.')
				this.$store.commit(LOADING, false)
				return
			}

			const credentials = {
				username: this.credentials.username,
				email: this.credentials.email,
				password: this.credentials.password,
			}

			this.$store.dispatch('auth/register', credentials)
		},
	},
}
</script>
