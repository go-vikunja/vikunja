<template>
	<div>
		<h2 class="title has-text-centered">{{ $t('user.auth.resetPassword') }}</h2>
		<div class="box">
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
						<x-button :to="{ name: 'user.login' }" type="secondary">
							{{ $t('user.auth.login') }}
						</x-button>
					</div>
				</div>
				<div class="notification is-danger" v-if="errorMsg">
					{{ errorMsg }}
				</div>
			</form>
			<div class="has-text-centered" v-if="isSuccess">
				<div class="notification is-success">
					{{ $t('user.auth.resetPasswordSuccess') }}
				</div>
				<x-button :to="{ name: 'user.login' }">
					{{ $t('user.auth.login') }}
				</x-button>
			</div>
			<Legal />
		</div>
	</div>
</template>

<script setup>
import {ref, reactive} from 'vue'
import { useI18n } from 'vue-i18n'

import Legal from '@/components/misc/legal'

import PasswordResetModel from '@/models/passwordReset'
import PasswordResetService from '@/services/passwordReset'
import { useTitle } from '@/composables/useTitle'

const { t } = useI18n()
useTitle(() => t('user.auth.resetPassword'))

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
	} catch(e) {
		errorMsg.value = e.response.data.message
	}
}
</script>

<style scoped>
.button {
	margin: 0 0.4rem 0 0;
}
</style>
