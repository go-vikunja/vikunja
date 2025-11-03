<template>
	<Card
		v-if="isLocalUser"
		:title="$t('user.settings.updateEmailTitle')"
	>
		<form @submit.prevent="updateEmail">
			<div class="field">
				<label
					class="label"
					for="newEmail"
				>{{ $t('user.settings.updateEmailNew') }}</label>
				<div class="control">
					<input
						id="newEmail"
						v-model="emailUpdate.newEmail"
						class="input"
						:placeholder="$t('user.auth.emailPlaceholder')"
						type="email"
						@keyup.enter="updateEmail"
					>
				</div>
			</div>
			<div class="field">
				<label
					class="label"
					for="currentPasswordEmail"
				>{{ $t('user.settings.currentPassword') }}</label>
				<div class="control">
					<input
						id="currentPasswordEmail"
						v-model="emailUpdate.password"
						class="input"
						:placeholder="$t('user.settings.currentPasswordPlaceholder')"
						type="password"
						@keyup.enter="updateEmail"
					>
				</div>
			</div>
		</form>

		<XButton
			:loading="emailUpdateService.loading"
			class="is-fullwidth mbs-4"
			@click="updateEmail"
		>
			{{ $t('misc.save') }}
		</XButton>
	</Card>
</template>


<script setup lang="ts">
import {reactive, computed, shallowReactive} from 'vue'
import {useI18n} from 'vue-i18n'

import EmailUpdateService from '@/services/emailUpdate'
import EmailUpdateModel from '@/models/emailUpdate'
import {success} from '@/message'
import {useTitle} from '@/composables/useTitle'
import {useAuthStore} from '@/stores/auth'

defineOptions({name: 'UserSettingsUpdateEmail'})

const {t} = useI18n({useScope: 'global'})
useTitle(() => `${t('user.settings.updateEmailTitle')} - ${t('user.settings.title')}`)

const authStore = useAuthStore()
const isLocalUser = computed(() => authStore.info?.isLocalUser)

const emailUpdate = reactive(new EmailUpdateModel())
const emailUpdateService = shallowReactive(new EmailUpdateService())
async function updateEmail() {
	await emailUpdateService.update(emailUpdate)
	success({message: t('user.settings.updateEmailSuccess')})
}
</script>
