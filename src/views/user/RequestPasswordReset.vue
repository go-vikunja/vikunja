<template>
	<div>
		<h2 class="title has-text-centered">Reset your password</h2>
		<div class="box">
			<form @submit.prevent="submit" v-if="!isSuccess">
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
							v-focus
							v-model="passwordReset.email"/>
					</div>
				</div>

				<div class="field is-grouped">
					<div class="control">
						<x-button
							@click="submit"
							:loading="passwordResetService.loading"
						>
							Send me a password reset link
						</x-button>
						<x-button :to="{ name: 'user.login' }" type="secondary">Login</x-button>
					</div>
				</div>
				<div class="notification is-danger" v-if="errorMsg">
					{{ errorMsg }}
				</div>
			</form>
			<div class="has-text-centered" v-if="isSuccess">
				<div class="notification is-success">
					Check your inbox! You should have a mail with instructions on how to reset your password.
				</div>
				<x-button :to="{ name: 'user.login' }">Login</x-button>
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
			passwordReset: PasswordResetModel,
			errorMsg: '',
			isSuccess: false,
		}
	},
	created() {
		this.passwordResetService = new PasswordResetService()
		this.passwordReset = new PasswordResetModel()
	},
	mounted() {
		this.setTitle('Reset your password')
	},
	methods: {
		submit() {
			this.errorMsg = ''
			this.passwordResetService.requestResetPassword(this.passwordReset)
				.then(() => {
					this.isSuccess = true
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
