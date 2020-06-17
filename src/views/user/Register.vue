<template>
	<div>
		<h2 class="title has-text-centered">Register</h2>
		<div class="box">
			<form id="registerform" @submit.prevent="submit">
				<div class="field">
					<label class="label" for="username">Username</label>
					<div class="control">
						<input v-focus type="text" id="username" class="input" name="username" placeholder="e.g. frederick" v-model="credentials.username" required/>
					</div>
				</div>
				<div class="field">
					<label class="label" for="email">E-mail address</label>
					<div class="control">
						<input type="email" class="input" id="email" name="email" placeholder="e.g. frederic@vikunja.io" v-model="credentials.email" required/>
					</div>
				</div>
				<div class="field">
					<label class="label" for="password1">Password</label>
					<div class="control">
						<input type="password" class="input" id="password1" name="password1" placeholder="e.g. ••••••••••••" v-model="credentials.password" required/>
					</div>
				</div>
				<div class="field">
					<label class="label" for="password2">Retype your password</label>
					<div class="control">
						<input type="password" class="input" id="password2" name="password2" placeholder="e.g. ••••••••••••" v-model="credentials.password2" required/>
					</div>
				</div>

				<div class="field is-grouped">
					<div class="control">
						<button type="submit" class="button is-primary" v-bind:class="{ 'is-loading': loading}">Register</button>
						<router-link :to="{ name: 'user.login' }" class="button">Login</router-link>
					</div>
				</div>
				<div class="notification is-info" v-if="loading">
					Loading...
				</div>
				<div class="notification is-danger" v-if="errorMessage !== ''">
					{{ errorMessage }}
				</div>
			</form>
		</div>
	</div>
</template>

<script>
	import router from '../../router'
	import {mapState} from 'vuex'
	import {ERROR_MESSAGE, LOADING} from '../../store/mutation-types'

	export default {
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
					password: this.credentials.password
				}

				this.$store.dispatch('auth/register', credentials)
			}
		}
	}
</script>

<style scoped>
	.button {
		margin: 0 0.4em 0 0;
	}
</style>
