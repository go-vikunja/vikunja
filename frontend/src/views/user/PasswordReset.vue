<template>
	<div>
		<Message
			v-if="errorMsg"
			class="mbe-4"
		>
			{{ errorMsg }}
		</Message>
		<div
			v-if="successMessage"
			class="has-text-centered mbe-4"
		>
			<Message variant="success">
				{{ successMessage }}
			</Message>
			<XButton
				:to="{ name: 'user.login' }"
				class="mbs-4"
			>
				{{ $t('user.auth.login') }}
			</XButton>
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
					<XButton
						:loading="passwordResetService.loading"
						@click="resetPassword"
					>
						{{ $t('user.auth.resetPassword') }}
					</XButton>
				</div>
			</div>
		</form>
	</div>
</template>

<script setup lang="ts">
import {ref, reactive} from 'vue'
import {useRoute} from 'vue-router'
import {useI18n} from 'vue-i18n'

import PasswordResetModel from '@/models/passwordReset'
import PasswordResetService from '@/services/passwordReset'
import Message from '@/components/misc/Message.vue'
import Password from '@/components/input/Password.vue'

const credentials = reactive({
	password: '',
})

const route = useRoute()
const {t} = useI18n()

const passwordResetService = reactive(new PasswordResetService())
const errorMsg = ref('')
const successMessage = ref('')

async function resetPassword() {
	errorMsg.value = ''
	const token = route.query.userPasswordReset as string

	if (!token) {
		errorMsg.value = t('user.auth.passwordResetTokenMissing')
		return
	}

	if (credentials.password === '') {
		return
	}

	const passwordReset = new PasswordResetModel({newPassword: credentials.password, token: token})
	try {
		const {message} = await passwordResetService.resetPassword(passwordReset)
		successMessage.value = message
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
