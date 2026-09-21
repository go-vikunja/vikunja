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
					v-model="credentials.password"
					@submit="resetPassword"
				/>
			</div>

			<div class="field is-grouped">
				<div class="control">
					<XButton
						:loading="passwordResetMutation.isPending.value"
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

import {usePasswordResetMutation} from '@/client/queries/passwords'
import Message from '@/components/misc/Message.vue'
import {getErrorText} from '@/message'
import Password from '@/components/input/Password.vue'

const credentials = reactive({
	password: '',
})

const route = useRoute()
const {t} = useI18n()

const passwordResetMutation = usePasswordResetMutation()
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

	try {
		const {message} = await passwordResetMutation.mutateAsync({new_password: credentials.password, token})
		successMessage.value = message ?? t('error.success')
	} catch (e) {
		errorMsg.value = getErrorText(e)
	}
}
</script>

<style scoped>
.button {
	margin: 0 0.4rem 0 0;
}
</style>
