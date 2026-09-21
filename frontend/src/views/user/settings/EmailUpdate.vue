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
				v-model="emailUpdate.new_email"
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

import {useUpdateEmailMutation, useCancelEmailUpdateMutation, useResendEmailConfirmationMutation} from '@/client/queries/email'
import FormField from '@/components/input/FormField.vue'
import Message from '@/components/misc/Message.vue'
import {useTitle} from '@/composables/useTitle'
import {useAuthStore} from '@/stores/auth'

defineOptions({name: 'UserSettingsUpdateEmail'})

const {t} = useI18n({useScope: 'global'})
useTitle(() => `${t('user.settings.updateEmailTitle')} - ${t('user.settings.title')}`)

const authStore = useAuthStore()
const isLocalUser = computed(() => authStore.info?.is_local_user)
const pendingEmail = computed(() => authStore.info?.pending_email)

const emailUpdate = reactive({new_email: '', password: ''})
const updateMutation = useUpdateEmailMutation()
const cancelMutation = useCancelEmailUpdateMutation()
const resendMutation = useResendEmailConfirmationMutation()
const pendingAction = ref<'save' | 'resend' | 'cancel' | null>(null)

async function runAction(name: 'save' | 'resend' | 'cancel', fn: () => Promise<void>) {
	if (pendingAction.value !== null) {
		return
	}
	pendingAction.value = name
	try {
		await fn()
	} catch {
		return
	} finally {
		pendingAction.value = null
	}
}

function updateEmail() {
	return runAction('save', async () => {
		await updateMutation.mutateAsync(emailUpdate)
		emailUpdate.new_email = ''
		emailUpdate.password = ''
	})
}

function resendConfirmation() {
	return runAction('resend', async () => {
		await resendMutation.mutateAsync()
	})
}

function cancelEmailUpdate() {
	return runAction('cancel', async () => {
		await cancelMutation.mutateAsync()
	})
}
</script>
