<template>
	<div>
		<Message
			v-if="errorMsg"
			class="mb-4"
		>
			{{ errorMsg }}
		</Message>
		<div
			v-if="successMessage"
			class="has-text-centered mb-4"
		>
			<Message variant="success">
				{{ successMessage }}
			</Message>
			<x-button
				:to="{ name: 'user.login' }"
				class="mt-4"
			>
				{{ $t('user.auth.login') }}
			</x-button>
		</div>
		<form
			v-if="!successMessage"
			id="form"
			@submit.prevent="resetPassword"
		>
			<div class="field">
				<label
					class="label"
					for="password"
				>{{ $t('user.auth.password') }}</label>
				<Password
					@submit="resetPassword"
					@update:modelValue="v => credentials.password = v"
				/>
			</div>

			<div class="field is-grouped">
				<div class="control">
					<x-button
						:loading="passwordResetService.loading"
						@click="resetPassword"
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

import PasswordResetModel from '@/models/passwordReset'
import PasswordResetService from '@/services/passwordReset'
import Message from '@/components/misc/Message.vue'
import Password from '@/components/input/Password.vue'

const credentials = reactive({
	password: '',
})

const passwordResetService = reactive(new PasswordResetService())
const errorMsg = ref('')
const successMessage = ref('')

async function resetPassword() {
	errorMsg.value = ''
	
	if(credentials.password === '') {
		return
	}

	const passwordReset = new PasswordResetModel({newPassword: credentials.password})
	try {
		const {message} = await passwordResetService.resetPassword(passwordReset)
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
