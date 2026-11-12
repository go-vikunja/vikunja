<template>
	<!-- The form seeds its draft once, so it may only mount with the account already loaded. -->
	<GeneralSettingsForm
		v-if="identity"
		ref="form"
		:key="identity.id"
		:initial-settings="authStore.settings"
		:is-external-user="authStore.info?.is_local_user === false"
		:auth-provider="authStore.info?.auth_provider"
		:saving="isSaving"
		own-account
		@save="save"
	/>
</template>

<script setup lang="ts">
import {computed, useTemplateRef} from 'vue'
import {useI18n} from 'vue-i18n'

import GeneralSettingsForm from './GeneralSettingsForm.vue'
import {useUpdateSettingsMutation, type UserSettingsEdits} from '@/client/queries/account'
import {useTitle} from '@/composables/useTitle'
import type {UserSettings} from '@/helpers/userSettings'
import {setLanguage} from '@/i18n'
import {error} from '@/message'
import {useAuthStore} from '@/stores/auth'
import {useConfigStore} from '@/stores/config'

defineOptions({name: 'UserSettingsGeneral'})

const {t} = useI18n({useScope: 'global'})
useTitle(() => `${t('user.settings.general.title')} - ${t('user.settings.title')}`)

const authStore = useAuthStore()
const configStore = useConfigStore()
const {mutate: updateUserSettings, isPending: isSaving} = useUpdateSettingsMutation()
const form = useTemplateRef('form')

const identity = computed(() => {
	const session = authStore.session
	if (!session || !authStore.currentAccount) return null
	return {
		id: session.id,
		type: session.type,
	}
})

function save(edits: UserSettingsEdits, submitted: UserSettings) {
	if (!identity.value) return
	if (configStore.demo_mode_enabled) delete edits.language
	// Captured now: a re-keyed form must not adopt this baseline.
	const target = form.value
	updateUserSettings({...identity.value, edits}, {
		onSuccess: () => {
			target?.markSaved(submitted)
			if (configStore.demo_mode_enabled) setLanguage(submitted.language).catch(error)
		},
	})
}
</script>
