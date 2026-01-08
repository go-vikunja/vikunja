<template>
	<Card
		v-if="isLocalUser"
		:title="$t('user.settings.newPasswordTitle')"
		:loading="passwordUpdateService.loading"
	>
		<form @submit.prevent="updatePassword">
			<FormField
				id="newPassword"
				v-model="passwordUpdate.newPassword"
				:label="$t('user.settings.newPassword')"
				autocomplete="new-password"
				:placeholder="$t('user.auth.passwordPlaceholder')"
				type="password"
				@keyup.enter="updatePassword"
			/>
			<FormField
				id="newPasswordConfirm"
				v-model="passwordConfirm"
				:label="$t('user.settings.newPasswordConfirm')"
				autocomplete="new-password"
				:placeholder="$t('user.auth.passwordPlaceholder')"
				type="password"
				@keyup.enter="updatePassword"
			/>
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
import FormField from '@/components/input/FormField.vue'

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
