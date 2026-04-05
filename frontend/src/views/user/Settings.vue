<template>
	<SideNavShell
		:navigation-items="navigationItems"
		:extra-links="extraSettingsLinks"
	/>
</template>

<script setup lang="ts">
import {computed} from 'vue'
import {useI18n} from 'vue-i18n'
import {useTitle} from '@/composables/useTitle'
import {useConfigStore} from '@/stores/config'
import {useAuthStore} from '@/stores/auth'

import SideNavShell from '@/components/misc/SideNavShell.vue'

const {t} = useI18n({useScope: 'global'})
useTitle(() => t('user.settings.title'))

const configStore = useConfigStore()
const authStore = useAuthStore()

const totpEnabled = computed(() => configStore.totpEnabled)
const caldavEnabled = computed(() => configStore.caldavEnabled)
const migratorsEnabled = computed(() => configStore.migratorsEnabled)
const isLocalUser = computed(() => authStore.info?.isLocalUser)
const userDeletionEnabled = computed(() => configStore.userDeletionEnabled)
const webhooksEnabled = computed(() => configStore.webhooksEnabled)
const botUsersEnabled = computed(() => configStore.botUsersEnabled)

const navigationItems = computed(() => {
	const items = [
		{
			title: t('user.settings.general.title'),
			routeName: 'user.settings.general',
		},
		{
			title: t('user.settings.newPasswordTitle'),
			routeName: 'user.settings.password-update',
			condition: isLocalUser.value,
		},
		{
			title: t('user.settings.updateEmailTitle'),
			routeName: 'user.settings.email-update',
			condition: isLocalUser.value,
		},
		{
			title: t('user.settings.avatar.title'),
			routeName: 'user.settings.avatar',
		},
		{
			title: t('user.settings.totp.title'),
			routeName: 'user.settings.totp',
			condition: totpEnabled.value && isLocalUser.value,
		},
		{
			title: t('user.export.title'),
			routeName: 'user.settings.data-export',
		},
		{
			title: t('migrate.title'),
			routeName: 'migrate.start',
			activeRouteNames: ['migrate.service'],
			condition: migratorsEnabled.value,
		},
		{
			title: t('user.settings.caldav.title'),
			routeName: 'user.settings.caldav',
			condition: caldavEnabled.value,
		},
		{
			title: t('user.settings.apiTokens.title'),
			routeName: 'user.settings.apiTokens',
		},
		{
			title: t('user.settings.sessions.title'),
			routeName: 'user.settings.sessions',
		},
		{
			title: t('user.settings.webhooks.title'),
			routeName: 'user.settings.webhooks',
			condition: webhooksEnabled.value,
		},
		{
			title: t('user.settings.bots.title'),
			routeName: 'user.settings.bots',
			condition: botUsersEnabled.value,
		},
		{
			title: t('user.deletion.title'),
			routeName: 'user.settings.deletion',
			condition: userDeletionEnabled.value,
		},
	]

	return items.filter(({condition}) => condition !== false)
})

const extraSettingsLinks = computed(() => Object.values(authStore.settings.extraSettingsLinks ?? {}))
</script>
