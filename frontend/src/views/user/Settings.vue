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

const totp_enabled = computed(() => configStore.totp_enabled)
const caldav_enabled = computed(() => configStore.caldav_enabled)
const migratorsEnabled = computed(() => configStore.migratorsEnabled)
const isLocalUser = computed(() => authStore.info?.is_local_user)
const user_deletion_enabled = computed(() => configStore.user_deletion_enabled)
const webhooks_enabled = computed(() => configStore.webhooks_enabled)

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
			condition: totp_enabled.value && isLocalUser.value,
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
			condition: caldav_enabled.value,
		},
		{
			title: 'MCP',
			routeName: 'user.settings.mcp',
		},
		{
			title: t('user.settings.feeds.title'),
			routeName: 'user.settings.feeds',
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
			condition: webhooks_enabled.value,
		},
		{
			title: t('user.settings.bots.title'),
			routeName: 'user.settings.bots',
		},
		{
			title: t('user.deletion.title'),
			routeName: 'user.settings.deletion',
			condition: user_deletion_enabled.value,
		},
	]

	return items.filter(({condition}) => condition !== false)
})

const extraSettingsLinks = computed(() => Object.values(authStore.settings.extraSettingsLinks ?? {}))
</script>
