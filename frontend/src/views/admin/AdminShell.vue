<template>
	<SideNavShell
		:navigation-items="navigationItems"
		exact
	/>
</template>

<script setup lang="ts">
import {computed} from 'vue'
import {useI18n} from 'vue-i18n'
import {useTitle} from '@/composables/useTitle'

import SideNavShell from '@/components/misc/SideNavShell.vue'
import {useConfigStore} from '@/stores/config'
import {PRO_FEATURE} from '@/constants/proFeatures'

const {t} = useI18n({useScope: 'global'})
useTitle(() => t('admin.title'))

const configStore = useConfigStore()

const navigationItems = computed(() => [
	{
		title: t('navigation.overview'),
		routeName: 'admin.overview',
	},
	{
		title: t('admin.labels.users'),
		routeName: 'admin.users',
	},
	{
		title: t('project.projects'),
		routeName: 'admin.projects',
	},
	...(configStore.isProFeatureEnabled(PRO_FEATURE.USER_INVITES) ? [{
		title: t('admin.inviteLinks.title'),
		routeName: 'admin.inviteLinks',
	}] : []),
])
</script>
