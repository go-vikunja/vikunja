<template>
	<card v-if="isLocalUser" :title="$t('user.settings.updateEmailTitle')">
		<form @submit.prevent="updateEmail">
			<div class="field">
				<label class="label" for="newEmail">{{ $t('user.settings.updateEmailNew') }}</label>
				<div class="control">
					<input
						@keyup.enter="updateEmail"
						class="input"
						id="newEmail"
						:placeholder="$t('user.auth.emailPlaceholder')"
						type="email"
						v-model="emailUpdate.newEmail"/>
				</div>
			</div>
			<div class="field">
				<label class="label" for="currentPasswordEmail">{{ $t('user.settings.currentPassword') }}</label>
				<div class="control">
					<input
						@keyup.enter="updateEmail"
						class="input"
						id="currentPasswordEmail"
						:placeholder="$t('user.settings.currentPasswordPlaceholder')"
						type="password"
						v-model="emailUpdate.password"/>
				</div>
			</div>
		</form>

		<x-button
			:loading="emailUpdateService.loading"
			@click="updateEmail"
			class="is-fullwidth mt-4">
			{{ $t('misc.save') }}
		</x-button>
	</card>
</template>

<script lang="ts">
import {defineComponent} from 'vue'
import EmailUpdateService from '@/services/emailUpdate'
import EmailUpdateModel from '@/models/emailUpdate'

export default defineComponent({
	name: 'user-settings-update-email',
	data() {
		return {
			emailUpdateService: new EmailUpdateService(),
			emailUpdate: new EmailUpdateModel(),
		}
	},
	mounted() {
		this.setTitle(`${this.$t('user.settings.updateEmailTitle')} - ${this.$t('user.settings.title')}`)
	},
	computed: {
		isLocalUser() {
			return this.$store.state.auth.info?.isLocalUser
		},
	},
	methods: {
		async updateEmail() {
			await this.emailUpdateService.update(this.emailUpdate)
			this.$message.success({message: this.$t('user.settings.updateEmailSuccess')})
		},
	},
})
</script>
