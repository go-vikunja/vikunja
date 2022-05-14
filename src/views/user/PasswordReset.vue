<template>
	<div>
		<message v-if="errorMsg" class="mb-4">
			{{ errorMsg }}
		</message>
		<div class="has-text-centered mb-4" v-if="successMessage">
			<message variant="success">
				{{ successMessage }}
			</message>
			<x-button :to="{ name: 'user.login' }">
				{{ $t('user.auth.login') }}
			</x-button>
		</div>
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
						v-model="credentials.password2"
						@keyup.enter="submit"
					/>
				</div>
			</div>

			<div class="field is-grouped">
				<div class="control">
					<x-button
						:loading="this.passwordResetService.loading"
						@click="submit"
					>
						{{ $t('user.auth.resetPassword') }}
					</x-button>
				</div>
			</div>
		</form>
	</div>
</template>

<script setup lang="ts">
import {ref, reactive} from 'vue'
import {useI18n} from 'vue-i18n'

import PasswordResetModel from '@/models/passwordReset'
import PasswordResetService from '@/services/passwordReset'
import Message from '@/components/misc/message.vue'

const {t} = useI18n()

const credentials = reactive({
	password: '',
	password2: '',
})

const passwordResetService = reactive(new PasswordResetService())
const errorMsg = ref('')
const successMessage = ref('')

async function submit() {
	errorMsg.value = ''

	if (credentials.password2 !== credentials.password) {
		errorMsg.value = t('user.auth.passwordsDontMatch')
		return
	}

	const passwordReset = new PasswordResetModel({newPassword: credentials.password})
	try {
		const {message} = passwordResetService.resetPassword(passwordReset)
		successMessage.value = message
		localStorage.removeItem('passwordResetToken')
	} catch (e) {
		errorMsg.value = e.response.data.message
	}
}
</script>

<style scoped>
.button {
	margin: 0 0.4rem 0 0;
}
</style>
