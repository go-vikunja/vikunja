<template>
	<div>
		<h2 class="title">Reset your password</h2>
		<div class="box">
			<form id="loginform" @submit.prevent="submit" v-if="!isSuccess">
				<div class="field">
					<div class="control">
						<input v-focus type="text" class="input" name="email" placeholder="Email-Adress" v-model="passwordReset.email" required>
					</div>
				</div>

				<div class="field is-grouped">
					<div class="control">
						<button type="submit" class="button is-primary" v-bind:class="{ 'is-loading': passwordResetService.loading}">Send me a password reset link</button>
						<router-link :to="{ name: 'login' }" class="button">Login</router-link>
					</div>
				</div>
				<div class="notification is-danger" v-if="error">
					{{ error }}
				</div>
			</form>
			<div v-if="isSuccess" class="has-text-centered">
				<div class="notification is-success">
					Check your inbox! You should have a mail with instructions on how to reset your password.
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
				passwordReset: PasswordResetModel,
				error: '',
				isSuccess: false
			}
		},
		created() {
			this.passwordResetService = new PasswordResetService()
			this.passwordReset = new PasswordResetModel()
		},
		methods: {
			submit() {
				this.error = ''
				this.passwordResetService.requestResetPassword(this.passwordReset)
					.then(() => {
						this.isSuccess = true
					})
					.catch(e => {
						this.error = e.response.data.message
					})
			},
		}
	}
</script>

<style scoped>
	.button {
		margin: 0 0.4em 0 0;
	}
</style>
