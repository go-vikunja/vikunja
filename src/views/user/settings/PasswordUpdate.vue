<template>
	<card v-if="isLocalUser" :title="$t('user.settings.newPasswordTitle')" :loading="passwordUpdateService.loading">
		<form @submit.prevent="updatePassword">
			<div class="field">
				<label class="label" for="newPassword">{{ $t('user.settings.newPassword') }}</label>
				<div class="control">
					<input
						autocomplete="new-password"
						@keyup.enter="updatePassword"
						class="input"
						id="newPassword"
						:placeholder="$t('user.auth.passwordPlaceholder')"
						type="password"
						v-model="passwordUpdate.newPassword"/>
				</div>
			</div>
			<div class="field">
				<label class="label" for="newPasswordConfirm">{{ $t('user.settings.newPasswordConfirm') }}</label>
				<div class="control">
					<input
						autocomplete="new-password"
						@keyup.enter="updatePassword"
						class="input"
						id="newPasswordConfirm"
						:placeholder="$t('user.auth.passwordPlaceholder')"
						type="password"
						v-model="passwordConfirm"/>
				</div>
			</div>
			<div class="field">
				<label class="label" for="currentPassword">{{ $t('user.settings.currentPassword') }}</label>
				<div class="control">
					<input
						autocomplete="current-password"
						@keyup.enter="updatePassword"
						class="input"
						id="currentPassword"
						:placeholder="$t('user.settings.currentPasswordPlaceholder')"
						type="password"
						v-model="passwordUpdate.oldPassword"/>
				</div>
			</div>
		</form>

		<x-button
			:loading="passwordUpdateService.loading"
			@click="updatePassword"
			class="is-fullwidth mt-4">
			{{ $t('misc.save') }}
		</x-button>
	</card>
</template>

<script>
import PasswordUpdateService from '@/services/passwordUpdateService'
import PasswordUpdateModel from '@/models/passwordUpdate'

export default {
	name: 'user-settings-password-update',
	data() {
		return {
			passwordUpdateService: new PasswordUpdateService(),
			passwordUpdate: new PasswordUpdateModel(),
			passwordConfirm: '',
		}
	},
	mounted() {
		this.setTitle(`${this.$t('user.settings.newPasswordTitle')} - ${this.$t('user.settings.title')}`)
	},
	computed: {
		isLocalUser() {
			return this.$store.state.auth.info?.isLocalUser
		},
	},
	methods: {
		async updatePassword() {
			if (this.passwordConfirm !== this.passwordUpdate.newPassword) {
				this.$message.error({message: this.$t('user.settings.passwordsDontMatch')})
				return
			}

			await this.passwordUpdateService.update(this.passwordUpdate)
			this.$message.success({message: this.$t('user.settings.passwordUpdateSuccess')})
		},
	},
}
</script>
