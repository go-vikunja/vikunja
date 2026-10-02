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

const {t} = useI18n({useScope: 'global'})
useTitle(() => t('manage.title'))

const navigationItems = computed(() => [
	{
		title: t('manage.users.title'),
		routeName: 'manage.users',
	},
	{
		title: t('manage.import.title'),
		routeName: 'manage.import',
	},
])
</script>
