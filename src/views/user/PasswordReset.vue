<template>
	<div>
		<h2 class="title has-text-centered">Reset your password</h2>
		<div class="box">
			<form @submit.prevent="submit" id="form" v-if="!successMessage">
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
							v-focus
							v-model="credentials.password"/>
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
							v-model="credentials.password2"/>
					</div>
				</div>

				<div class="field is-grouped">
					<div class="control">
						<x-button
							:loading="this.passwordResetService.loading"
							@click="submit"
						>
							Reset your password
						</x-button>
					</div>
				</div>
				<div class="notification is-info" v-if="this.passwordResetService.loading">
					Loading...
				</div>
				<div class="notification is-danger" v-if="errorMsg">
					{{ errorMsg }}
				</div>
			</form>
			<div class="has-text-centered" v-if="successMessage">
				<div class="notification is-success">
					{{ successMessage }}
				</div>
				<x-button :to="{ name: 'user.login' }">
					Login
				</x-button>
			</div>
			<legal/>
		</div>
	</div>
</template>

<script>
import PasswordResetModel from '../../models/passwordReset'
import PasswordResetService from '../../services/passwordReset'
import Legal from '../../components/misc/legal'

export default {
	components: {
		Legal,
	},
	data() {
		return {
			passwordResetService: PasswordResetService,
			credentials: {
				password: '',
				password2: '',
			},
			errorMsg: '',
			successMessage: '',
		}
	},
	created() {
		this.passwordResetService = new PasswordResetService()
	},
	mounted() {
		this.setTitle('Reset your password')
	},
	methods: {
		submit() {
			this.errorMsg = ''

			if (this.credentials.password2 !== this.credentials.password) {
				this.errorMsg = 'Passwords don\'t match'
				return
			}

			let passwordReset = new PasswordResetModel({newPassword: this.credentials.password})
			this.passwordResetService.resetPassword(passwordReset)
				.then(response => {
					this.successMessage = response.message
					localStorage.removeItem('passwordResetToken')
				})
				.catch(e => {
					this.errorMsg = e.response.data.message
				})
		},
	},
}
</script>

<style scoped>
.button {
	margin: 0 0.4em 0 0;
}
</style>
