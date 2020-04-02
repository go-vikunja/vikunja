<template>
	<div>
		<h2 class="title has-text-centered">Login</h2>
		<div class="box">
			<div v-if="confirmedEmailSuccess" class="notification is-success has-text-centered">
				You successfully confirmed your email! You can log in now.
			</div>
			<form id="loginform" @submit.prevent="submit">
				<div class="field">
					<label class="label" for="username">Username</label>
					<div class="control">
						<input v-focus type="text" id="username" class="input" name="username" placeholder="e.g. frederick" ref="username" required/>
					</div>
				</div>
				<div class="field">
					<label class="label" for="password">Password</label>
					<div class="control">
						<input type="password" class="input" id="password" name="password" placeholder="e.g. ••••••••••••" ref="password" required/>
					</div>
				</div>

				<div class="field is-grouped">
					<div class="control">
						<button type="submit" class="button is-primary" v-bind:class="{ 'is-loading': loading}">Login</button>
						<router-link :to="{ name: 'register' }" class="button">Register</router-link>
						<router-link :to="{ name: 'getPasswordReset' }" class="reset-password-link">Reset your password</router-link>
					</div>
				</div>
				<div class="notification is-danger" v-if="errorMsg">
					{{ errorMsg }}
				</div>
			</form>
		</div>
	</div>
</template>

<script>
	import auth from '../../auth'
	import router from '../../router'
	import {HTTP} from '../../http-common'
	import message from '../../message'

	export default {
		data() {
			return {
				errorMsg: '',
				confirmedEmailSuccess: false,
				loading: false
			}
		},
		beforeMount() {
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
						this.errorMsg = e.response.data.message
					})
			}

			// Check if the user is already logged in, if so, redirect him to the homepage
			if (auth.user.authenticated) {
				router.push({name: 'home'})
			}
		},
		methods: {
			submit() {
				this.loading = true
				this.errorMsg = ''
				// Some browsers prevent Vue bindings from working with autofilled values.
				// To work around this, we're manually getting the values here instead of relying on vue bindings.
				// For more info, see https://kolaente.dev/vikunja/frontend/issues/78
				const credentials = {
					username: this.$refs.username.value,
					password: this.$refs.password.value,
				}

				auth.login(this, credentials, 'home')
			}
		}
	}
</script>

<style scoped>
	.button {
		margin: 0 0.4em 0 0;
	}

	.reset-password-link{
		display: inline-block;
		padding-top: 5px;
	}
</style>
