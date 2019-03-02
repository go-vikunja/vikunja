<template>
	<div>
		<h2 class="title">Reset your password</h2>
		<div class="box">
			<form id="form" @submit.prevent="submit" v-if="!successMessage">
				<div class="field">
					<div class="control">
						<input v-focus type="password" class="input" name="password1" placeholder="Password" v-model="credentials.password" required>
					</div>
				</div>
				<div class="field">
					<div class="control">
						<input type="password" class="input" name="password2" placeholder="Retype password" v-model="credentials.password2" required>
					</div>
				</div>

				<div class="field is-grouped">
					<div class="control">
						<button type="submit" class="button is-primary" :class="{ 'is-loading': this.passwordResetService.loading}">Reset your password</button>
					</div>
				</div>
				<div class="notification is-info" v-if="this.passwordResetService.loading">
					Loading...
				</div>
				<div class="notification is-danger" v-if="error">
					{{ error }}
				</div>
			</form>
			<div v-if="successMessage" class="has-text-centered">
				<div class="notification is-success">
					{{ successMessage }}
				</div>
				<router-link :to="{ name: 'login' }" class="button is-primary">Login</router-link>
			</div>
		</div>
	</div>
</template>

<script>
	import PasswordResetModel from '../../models/passwordReset'
	import PasswordResetService from '../../services/passwordReset'

	export default {
		data() {
			return {
				passwordResetService: PasswordResetService,
				credentials: {
					password: '',
					password2: '',
				},
				error: '',
				successMessage: ''
			}
		},
		created() {
			this.passwordResetService = new PasswordResetService()
		},
		methods: {
			submit() {
				this.error = ''

				if (this.credentials.password2 !== this.credentials.password) {
					this.error = 'Passwords don\'t match'
					return
				}

				let passwordReset = new PasswordResetModel({new_password: this.credentials.password})
				this.passwordResetService.resetPassword(passwordReset)
					.then(response => {
						this.successMessage = response.data.message
						localStorage.removeItem('passwordResetToken')
					})
					.catch(e => {
						this.error = e.response.data.message
					})
			}
		}
	}
</script>

<style scoped>
	.button {
		margin: 0 0.4em 0 0;
	}
</style>
