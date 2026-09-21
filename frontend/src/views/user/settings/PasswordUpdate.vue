<template>
	<Card
		v-if="isLocalUser"
		:title="$t('user.settings.newPasswordTitle')"
		:loading="passwordUpdateMutation.isPending.value"
	>
		<form @submit.prevent="updatePassword">
			<div class="field">
				<label
					class="label"
					for="password"
				>{{ $t('user.settings.newPassword') }}</label>
				<Password
					v-model="passwordUpdate.new_password"
					:validate-initially="true"
					@submit="updatePassword"
				/>
			</div>
			<FormField
				id="currentPassword"
				v-model="passwordUpdate.old_password"
				:label="$t('user.settings.currentPassword')"
				autocomplete="current-password"
				:placeholder="$t('user.settings.currentPasswordPlaceholder')"
				type="password"
				@keyup.enter="updatePassword"
			/>
		</form>

		<XButton
			:loading="passwordUpdateMutation.isPending.value"
			:disabled="!isValid"
			class="is-fullwidth mbs-4"
			@click="updatePassword"
		>
			{{ $t('misc.save') }}
		</XButton>
	</Card>
</template>


<script setup lang="ts">
import {reactive, computed} from 'vue'
import {useI18n} from 'vue-i18n'

import {useChangePasswordMutation} from '@/client/queries/passwords'
import FormField from '@/components/input/FormField.vue'
import Password from '@/components/input/Password.vue'

import {useTitle} from '@/composables/useTitle'
import {useAuthStore} from '@/stores/auth'
import {validatePassword} from '@/helpers/validatePasswort'

defineOptions({name: 'UserSettingsPasswordUpdate'})

const passwordUpdateMutation = useChangePasswordMutation()
const passwordUpdate = reactive({new_password: '', old_password: ''})

const {t} = useI18n({useScope: 'global'})
useTitle(() => `${t('user.settings.newPasswordTitle')} - ${t('user.settings.title')}`)

const authStore = useAuthStore()
const isLocalUser = computed(() => authStore.info?.is_local_user)
const isValid = computed(() => validatePassword(passwordUpdate.new_password) === true && passwordUpdate.old_password !== '')

async function updatePassword() {
	try {
		await passwordUpdateMutation.mutateAsync(passwordUpdate)
		passwordUpdate.new_password = ''
		passwordUpdate.old_password = ''
	} catch { return }
}
</script>
