<template>
	<div>
		<h2 class="title has-text-centered">{{ $t('user.auth.resetPassword') }}</h2>
		<div class="box">
			<form @submit.prevent="submit" id="form" v-if="!successMessage">
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
							v-focus
							v-model="credentials.password"/>
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
							v-model="credentials.password2"/>
					</div>
				</div>

				<div class="field is-grouped">
					<div class="control">
						<x-button
							:loading="this.passwordResetService.loading"
							@click="submit"
						>
							{{ $t('user.auth.resetPassoword') }}
						</x-button>
					</div>
				</div>
				<div class="notification is-info" v-if="this.passwordResetService.loading">
					{{ $t('misc.loading') }}
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
					{{ $t('user.auth.login') }}
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
		this.setTitle(this.$t('user.auth.resetPassword'))
	},
	methods: {
		submit() {
			this.errorMsg = ''

			if (this.credentials.password2 !== this.credentials.password) {
				this.errorMsg = this.$t('user.auth.passwordsDontMatch')
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
	margin: 0 0.4rem 0 0;
}
</style>
