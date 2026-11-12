<template>
	<div class="admin-user-settings">
		<div class="admin-user-settings__header">
			<RouterLink :to="{name: 'admin.users'}">
				{{ $t('admin.users.backToUsers') }}
			</RouterLink>
			<h2
				v-if="data"
				class="title is-4"
			>
				{{ $t('admin.users.settingsTitle', {username: data.user?.username}) }}
			</h2>
		</div>
		<Message
			v-if="data && data.user?.id !== authStore.info?.id"
			variant="warning"
			class="admin-user-settings__warning"
		>
			{{ $t('admin.users.settingsEditingOther', {username: data.user?.username}) }}
		</Message>

		<p v-if="isPending">
			{{ $t('misc.loading') }}
		</p>
		<!-- The form seeds its draft once, so it may only mount with the settings loaded. -->
		<GeneralSettingsForm
			v-else-if="data"
			ref="form"
			:key="userId"
			:initial-settings="data.settings"
			:is-external-user="!!data.user?.auth_provider"
			:auth-provider="data.user?.auth_provider"
			:saving="isSaving"
			:own-account="false"
			@save="save"
		/>
	</div>
</template>

<script setup lang="ts">
import {computed, useTemplateRef} from 'vue'
import {useI18n} from 'vue-i18n'
import {useQuery} from '@tanstack/vue-query'

import Message from '@/components/misc/Message.vue'
import GeneralSettingsForm from '@/views/user/settings/GeneralSettingsForm.vue'
import type {UserSettingsEdits} from '@/client/queries/account'
import type {UserSettings} from '@/helpers/userSettings'
import {adminUserSettingsQuery, useUpdateAdminUserSettingsMutation} from '@/client/queries/admin'
import {useTitle} from '@/composables/useTitle'
import {useAuthStore} from '@/stores/auth'

const props = defineProps<{userId: number}>()

defineOptions({name: 'AdminUserSettings'})

const {t} = useI18n({useScope: 'global'})
const authStore = useAuthStore()

const {data, isPending} = useQuery(computed(() => adminUserSettingsQuery(props.userId)))
useTitle(() => data.value && t('admin.users.settingsTitle', {username: data.value.user?.username}))
const {mutate: updateSettings, isPending: isSaving} = useUpdateAdminUserSettingsMutation()
const form = useTemplateRef('form')

function save(edits: UserSettingsEdits, submitted: UserSettings) {
	// Captured now: a re-keyed form must not adopt this baseline.
	const target = form.value
	updateSettings({id: props.userId, edits}, {
		onSuccess: () => target?.markSaved(submitted),
	})
}
</script>

<style lang="scss" scoped>
.admin-user-settings__header {
	margin-block-end: 1rem;

	.title {
		margin-block-start: .5rem;
	}
}

// Not inside the header: sticky only sticks within its parent.
.admin-user-settings__warning {
	position: sticky;
	inset-block-start: $navbar-height;
	z-index: 5;
	margin-block-end: 1rem;
}
</style>
