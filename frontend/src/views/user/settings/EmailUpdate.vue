<template>
	<Card
		v-if="isLocalUser"
		:title="$t('user.settings.updateEmailTitle')"
	>
		<Message
			v-if="pendingEmail"
			variant="warning"
			role="status"
			class="mbe-4"
		>
			<p>{{ $t('user.settings.updateEmailPending', {email: pendingEmail}) }}</p>
			<div class="buttons mbs-2">
				<XButton
					variant="secondary"
					:loading="pendingAction === 'resend'"
					:disabled="pendingAction !== null"
					@click="resendConfirmation"
				>
					{{ $t('user.settings.updateEmailResend') }}
				</XButton>
				<XButton
					variant="tertiary"
					:loading="pendingAction === 'cancel'"
					:disabled="pendingAction !== null"
					@click="cancelEmailUpdate"
				>
					{{ $t('user.settings.updateEmailCancel') }}
				</XButton>
			</div>
		</Message>
		<form @submit.prevent="updateEmail">
			<FormField
				id="newEmail"
				v-model="emailUpdate.newEmail"
				:label="$t('user.settings.updateEmailNew')"
				:placeholder="$t('user.auth.emailPlaceholder')"
				type="email"
				name="email"
				autocomplete="email"
				@keyup.enter="updateEmail"
			/>
			<FormField
				id="currentPasswordEmail"
				v-model="emailUpdate.password"
				:label="$t('user.settings.currentPassword')"
				:placeholder="$t('user.settings.currentPasswordPlaceholder')"
				type="password"
				name="current-password"
				autocomplete="current-password"
				@keyup.enter="updateEmail"
			/>
		</form>

		<XButton
			:loading="pendingAction === 'save'"
			:disabled="pendingAction !== null"
			class="is-fullwidth mbs-4"
			@click="updateEmail"
		>
			{{ $t('misc.save') }}
		</XButton>
	</Card>
</template>


<script setup lang="ts">
import {reactive, computed, ref} from 'vue'
import {useI18n} from 'vue-i18n'

import EmailUpdateService from '@/services/emailUpdate'
import EmailUpdateModel from '@/models/emailUpdate'
import FormField from '@/components/input/FormField.vue'
import Message from '@/components/misc/Message.vue'
import {success} from '@/message'
import {useTitle} from '@/composables/useTitle'
import {useAuthStore} from '@/stores/auth'

defineOptions({name: 'UserSettingsUpdateEmail'})

const {t} = useI18n({useScope: 'global'})
useTitle(() => `${t('user.settings.updateEmailTitle')} - ${t('user.settings.title')}`)

const authStore = useAuthStore()
const isLocalUser = computed(() => authStore.info?.isLocalUser)
const pendingEmail = computed(() => authStore.info?.pendingEmail)

const emailUpdate = reactive(new EmailUpdateModel())
const emailUpdateService = new EmailUpdateService()
const pendingAction = ref<'save' | 'resend' | 'cancel' | null>(null)

async function runAction(name: 'save' | 'resend' | 'cancel', fn: () => Promise<void>) {
	if (pendingAction.value !== null) {
		return
	}
	pendingAction.value = name
	try {
		await fn()
	} finally {
		pendingAction.value = null
	}
}

function updateEmail() {
	return runAction('save', async () => {
		await emailUpdateService.update(emailUpdate)
		emailUpdate.newEmail = ''
		emailUpdate.password = ''
		await authStore.refreshUserInfo()
		success({message: authStore.info?.pendingEmail
			? t('user.settings.updateEmailPendingSuccess')
			: t('user.settings.updateEmailSuccess'),
		})
	})
}

function resendConfirmation() {
	return runAction('resend', async () => {
		await emailUpdateService.resend()
		success({message: t('user.settings.updateEmailResendSuccess')})
	})
}

function cancelEmailUpdate() {
	return runAction('cancel', async () => {
		await emailUpdateService.cancel()
		success({message: t('user.settings.updateEmailCancelSuccess')})
		await authStore.refreshUserInfo()
	})
}
</script>
