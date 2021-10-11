<template>
	<div>
		<h2 class="title has-text-centered">{{ $t('user.auth.resetPassword') }}</h2>
		<div class="box">
			<form @submit.prevent="submit" v-if="!isSuccess">
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
							{{ $t('user.auth.resetPasswordAction') }}
						</x-button>
						<x-button :to="{ name: 'user.login' }" type="secondary">
							{{ $t('user.auth.login') }}
						</x-button>
					</div>
				</div>
				<div class="notification is-danger" v-if="errorMsg">
					{{ errorMsg }}
				</div>
			</form>
			<div class="has-text-centered" v-if="isSuccess">
				<div class="notification is-success">
					{{ $t('user.auth.resetPasswordSuccess') }}
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
			passwordResetService: new PasswordResetService(),
			passwordReset: new PasswordResetModel(),
			errorMsg: '',
			isSuccess: false,
		}
	},
	mounted() {
		this.setTitle(this.$t('user.auth.resetPassword'))
	},
	methods: {
		async submit() {
			this.errorMsg = ''
			try {
				await this.passwordResetService.requestResetPassword(this.passwordReset)
				this.isSuccess = true
			} catch(e) {
				this.errorMsg = e.response.data.message
			}
		},
	},
}
</script>

<style scoped>
.button {
	margin: 0 0.4rem 0 0;
}
</style>
