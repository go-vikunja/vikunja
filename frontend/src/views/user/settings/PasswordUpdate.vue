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
					for="password"
				>{{ $t('user.settings.newPassword') }}</label>
				<Password
					:validate-initially="true"
					@update:modelValue="v => passwordUpdate.newPassword = v"
					@submit="updatePassword"
				/>
			</div>
			<FormField
				id="currentPassword"
				v-model="passwordUpdate.oldPassword"
				:label="$t('user.settings.currentPassword')"
				autocomplete="current-password"
				:placeholder="$t('user.settings.currentPasswordPlaceholder')"
				type="password"
				@keyup.enter="updatePassword"
			/>
		</form>

		<XButton
			:loading="passwordUpdateService.loading"
			:disabled="!isValid"
			class="is-fullwidth mbs-4"
			@click="updatePassword"
		>
			{{ $t('misc.save') }}
		</XButton>
	</Card>
</template>


<script setup lang="ts">
import {reactive, shallowReactive, computed} from 'vue'
import {useI18n} from 'vue-i18n'

import PasswordUpdateService from '@/services/passwordUpdateService'
import PasswordUpdateModel from '@/models/passwordUpdate'
import FormField from '@/components/input/FormField.vue'
import Password from '@/components/input/Password.vue'

import {useTitle} from '@/composables/useTitle'
import {success} from '@/message'
import {useAuthStore} from '@/stores/auth'
import {validatePassword} from '@/helpers/validatePasswort'

defineOptions({name: 'UserSettingsPasswordUpdate'})

const passwordUpdateService = shallowReactive(new PasswordUpdateService())
const passwordUpdate = reactive(new PasswordUpdateModel())

const {t} = useI18n({useScope: 'global'})
useTitle(() => `${t('user.settings.newPasswordTitle')} - ${t('user.settings.title')}`)

const authStore = useAuthStore()
const isLocalUser = computed(() => authStore.info?.isLocalUser)
const isValid = computed(() => validatePassword(passwordUpdate.newPassword) === true && passwordUpdate.oldPassword !== '')

async function updatePassword() {
	await passwordUpdateService.update(passwordUpdate)
	success({message: t('user.settings.passwordUpdateSuccess')})
}
</script>
