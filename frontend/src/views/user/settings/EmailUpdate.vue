<template>
	<Card
		v-if="isLocalUser"
		:title="$t('user.settings.updateEmailTitle')"
	>
		<form @submit.prevent="updateEmail">
			<FormField
				id="newEmail"
				v-model="emailUpdate.newEmail"
				:label="$t('user.settings.updateEmailNew')"
				:placeholder="$t('user.auth.emailPlaceholder')"
				type="email"
				@keyup.enter="updateEmail"
			/>
			<FormField
				id="currentPasswordEmail"
				v-model="emailUpdate.password"
				:label="$t('user.settings.currentPassword')"
				:placeholder="$t('user.settings.currentPasswordPlaceholder')"
				type="password"
				@keyup.enter="updateEmail"
			/>
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
import FormField from '@/components/input/FormField.vue'
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
