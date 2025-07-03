<template>
	<div>
		<Message
			v-if="errorMsg"
			variant="danger"
			class="mb-4"
		>
			{{ errorMsg }}
		</Message>
		<div
			v-if="isSuccess"
			class="has-text-centered mb-4"
		>
			<Message variant="success">
				{{ $t('user.auth.resetPasswordSuccess') }}
			</Message>
			<XButton
				:to="{ name: 'user.login' }"
				class="mt-4"
			>
				{{ $t('user.auth.login') }}
			</XButton>
		</div>
		<form
			v-if="!isSuccess"
			@submit.prevent="requestPasswordReset"
		>
			<div class="field">
				<label
					class="label"
					for="email"
				>{{ $t('user.auth.email') }}</label>
				<div class="control">
					<input
						id="email"
						v-model="passwordReset.email"
						v-focus
						class="input"
						name="email"
						:placeholder="$t('user.auth.emailPlaceholder')"
						required
						type="email"
					>
				</div>
			</div>

			<div class="is-flex">
				<XButton
					type="submit"
					:loading="passwordResetService.loading"
				>
					{{ $t('user.auth.resetPasswordAction') }}
				</XButton>
				<XButton
					:to="{ name: 'user.login' }"
					variant="secondary"
				>
					{{ $t('user.auth.login') }}
				</XButton>
			</div>
		</form>
	</div>
</template>

<script setup lang="ts">
import {ref, shallowReactive} from 'vue'
import {useI18n} from 'vue-i18n'

import PasswordResetModel from '@/models/passwordReset'
import PasswordResetService from '@/services/passwordReset'
import Message from '@/components/misc/Message.vue'

const {t} = useI18n({useScope: 'global'})

const passwordResetService = shallowReactive(new PasswordResetService())
const passwordReset = ref(new PasswordResetModel())
const errorMsg = ref('')
const isSuccess = ref(false)

async function requestPasswordReset() {
	errorMsg.value = ''
	try {
		await passwordResetService.requestResetPassword(passwordReset.value)
		isSuccess.value = true
	} catch (e: unknown) {
		errorMsg.value = (e instanceof Error && 'response' in e ? (e as {response: {data: {message: string}}}).response.data.message : String(e)) || t('user.auth.requestPasswordResetError')
	}
}
</script>

<style scoped>
.button {
	margin: 0 0.4rem 0 0;
}
</style>
