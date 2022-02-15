<template>
	<div>
		<message variant="danger" v-if="errorMsg" class="mb-4">
			{{ errorMsg }}
		</message>
		<div class="has-text-centered mb-4" v-if="isSuccess">
			<message variant="success">
				{{ $t('user.auth.resetPasswordSuccess') }}
			</message>
			<x-button :to="{ name: 'user.login' }">
				{{ $t('user.auth.login') }}
			</x-button>
		</div>
		<form @submit.prevent="submit" v-if="!isSuccess">
			<div class="field">
				<label class="label" for="email">{{ $t('user.auth.email') }}</label>
				<div class="control">
					<input
						class="input"
						id="email"
						name="email"
						:placeholder="$t('user.auth.emailPlaceholder')"
						required
						type="email"
						v-focus
						v-model="passwordReset.email"/>
				</div>
			</div>

			<div class="field is-grouped">
				<div class="control">
					<x-button
						@click="submit"
						:loading="passwordResetService.loading"
					>
						{{ $t('user.auth.resetPasswordAction') }}
					</x-button>
					<x-button :to="{ name: 'user.login' }" variant="secondary">
						{{ $t('user.auth.login') }}
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
import Message from '@/components/misc/message'

// Not sure if this instance needs a shalloRef at all
const passwordResetService = reactive(new PasswordResetService())
const passwordReset = ref(new PasswordResetModel())
const errorMsg = ref('')
const isSuccess = ref(false)

async function submit() {
	errorMsg.value = ''
	try {
		await passwordResetService.requestResetPassword(passwordReset.value)
		isSuccess.value = true
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
