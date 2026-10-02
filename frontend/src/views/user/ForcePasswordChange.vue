<template>
	<div class="force-password-change">
		<Message
			variant="warning"
			class="mbe-4"
		>
			{{ $t('user.settings.forcePasswordChange.intro') }}
		</Message>
		<Message
			v-if="errorMsg"
			class="mbe-4"
		>
			{{ errorMsg }}
		</Message>

		<form @submit.prevent="changePassword">
			<FormField
				id="currentPassword"
				v-model="form.oldPassword"
				:label="$t('user.settings.currentPassword')"
				autocomplete="current-password"
				:placeholder="$t('user.settings.currentPasswordPlaceholder')"
				type="password"
				@keyup.enter="changePassword"
			/>
			<div class="field">
				<label
					class="label"
					for="password"
				>{{ $t('user.settings.newPassword') }}</label>
				<Password
					:validate-initially="true"
					@update:modelValue="v => form.newPassword = v"
					@submit="changePassword"
				/>
			</div>
			<p
				v-if="form.newPassword !== '' && form.newPassword === form.oldPassword"
				class="help is-danger"
			>
				{{ $t('user.settings.forcePasswordChange.mustDiffer') }}
			</p>
		</form>

		<div class="buttons mbs-4">
			<XButton
				:loading="saving"
				:disabled="!isValid"
				@click="changePassword"
			>
				{{ $t('misc.save') }}
			</XButton>
			<XButton
				variant="tertiary"
				:disabled="saving"
				@click="authStore.logout()"
			>
				{{ $t('user.auth.logout') }}
			</XButton>
		</div>
	</div>
</template>

<script setup lang="ts">
import {computed, reactive, ref} from 'vue'
import {useI18n} from 'vue-i18n'

import Message from '@/components/misc/Message.vue'
import FormField from '@/components/input/FormField.vue'
import Password from '@/components/input/Password.vue'
import XButton from '@/components/input/Button.vue'

import PasswordUpdateService from '@/services/passwordUpdateService'
import PasswordUpdateModel from '@/models/passwordUpdate'
import {validatePassword} from '@/helpers/validatePasswort'
import {getErrorText, success} from '@/message'
import {useAuthStore} from '@/stores/auth'
import {useTitle} from '@/composables/useTitle'

defineOptions({name: 'UserForcePasswordChange'})

const {t} = useI18n({useScope: 'global'})
useTitle(() => t('user.settings.forcePasswordChange.title'))

const authStore = useAuthStore()
const service = new PasswordUpdateService()
const form = reactive(new PasswordUpdateModel())
const saving = ref(false)
const errorMsg = ref('')

const isValid = computed(() =>
	validatePassword(form.newPassword) === true &&
	form.oldPassword !== '' &&
	form.newPassword !== form.oldPassword,
)

async function changePassword() {
	if (!isValid.value || saving.value) {
		return
	}

	saving.value = true
	errorMsg.value = ''
	try {
		await service.update(form)
		// The server ended every session of this user when the password changed, so the token in
		// this tab is useless now. Sign in again with the new password.
		success({message: t('user.settings.forcePasswordChange.success')})
		await authStore.logout()
	} catch (e) {
		errorMsg.value = getErrorText(e)
	} finally {
		saving.value = false
	}
}
</script>
