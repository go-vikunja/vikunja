<template>
	<Card
		v-if="isLocalUser"
		:title="$t('user.settings.newPasswordTitle')"
		:loading="passwordUpdateService.loading"
	>
		<form @submit.prevent="updatePassword">
			<div class="field">
				<label
					class="label"
					for="newPassword"
				>{{ $t('user.settings.newPassword') }}</label>
				<div class="control">
					<input
						id="newPassword"
						v-model="passwordUpdate.newPassword"
						autocomplete="new-password"
						class="input"
						:placeholder="$t('user.auth.passwordPlaceholder')"
						type="password"
						@keyup.enter="updatePassword"
					>
				</div>
			</div>
			<div class="field">
				<label
					class="label"
					for="newPasswordConfirm"
				>{{ $t('user.settings.newPasswordConfirm') }}</label>
				<div class="control">
					<input
						id="newPasswordConfirm"
						v-model="passwordConfirm"
						autocomplete="new-password"
						class="input"
						:placeholder="$t('user.auth.passwordPlaceholder')"
						type="password"
						@keyup.enter="updatePassword"
					>
				</div>
			</div>
			<div class="field">
				<label
					class="label"
					for="currentPassword"
				>{{ $t('user.settings.currentPassword') }}</label>
				<div class="control">
					<input
						id="currentPassword"
						v-model="passwordUpdate.oldPassword"
						autocomplete="current-password"
						class="input"
						:placeholder="$t('user.settings.currentPasswordPlaceholder')"
						type="password"
						@keyup.enter="updatePassword"
					>
				</div>
			</div>
		</form>

		<XButton
			:loading="passwordUpdateService.loading"
			class="is-fullwidth mbs-4"
			@click="updatePassword"
		>
			{{ $t('misc.save') }}
		</XButton>
	</Card>
</template>


<script setup lang="ts">
import {ref, reactive, shallowReactive, computed} from 'vue'
import {useI18n} from 'vue-i18n'

import PasswordUpdateService from '@/services/passwordUpdateService'
import PasswordUpdateModel from '@/models/passwordUpdate'

import {useTitle} from '@/composables/useTitle'
import {success, error} from '@/message'
import {useAuthStore} from '@/stores/auth'

defineOptions({name: 'UserSettingsPasswordUpdate'})

const passwordUpdateService = shallowReactive(new PasswordUpdateService())
const passwordUpdate = reactive(new PasswordUpdateModel())
const passwordConfirm = ref('')

const {t} = useI18n({useScope: 'global'})
useTitle(() => `${t('user.settings.newPasswordTitle')} - ${t('user.settings.title')}`)

const authStore = useAuthStore()
const isLocalUser = computed(() => authStore.info?.isLocalUser)

async function updatePassword() {
	if (passwordConfirm.value !== passwordUpdate.newPassword) {
		error({message: t('user.settings.passwordsDontMatch')})
		return
	}

	await passwordUpdateService.update(passwordUpdate)
	success({message: t('user.settings.passwordUpdateSuccess')})
}
</script>
