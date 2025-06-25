<template>
	<Ready>
		<template v-if="authStore.authUser">
			<AppHeader />
			<ContentAuth />
		</template>
		<ContentLinkShare v-else-if="authStore.authLinkShare" />
		<NoAuthWrapper
			v-else
			show-api-config
		>
			<RouterView />
		</NoAuthWrapper>
		
		<KeyboardShortcuts v-if="keyboardShortcutsActive" />
		
		<Teleport to="body">
			<AddToHomeScreen />
			<UpdateNotification />
			<Notification />
			<DemoMode />
		</Teleport>
	</Ready>
</template>

<script lang="ts" setup>
import {computed, watch} from 'vue'
import {useRoute, useRouter} from 'vue-router'
import {useI18n} from 'vue-i18n'
import isTouchDevice from 'is-touch-device'

import Notification from '@/components/misc/Notification.vue'
import UpdateNotification from '@/components/home/UpdateNotification.vue'
import KeyboardShortcuts from '@/components/misc/keyboard-shortcuts/index.vue'

import AppHeader from '@/components/home/AppHeader.vue'
import ContentAuth from '@/components/home/ContentAuth.vue'
import ContentLinkShare from '@/components/home/ContentLinkShare.vue'
import NoAuthWrapper from '@/components/misc/NoAuthWrapper.vue'
import Ready from '@/components/misc/Ready.vue'

import {setLanguage} from '@/i18n'

import {useAuthStore} from '@/stores/auth'
import {useBaseStore} from '@/stores/base'

import {useColorScheme} from '@/composables/useColorScheme'
import {useBodyClass} from '@/composables/useBodyClass'
import AddToHomeScreen from '@/components/home/AddToHomeScreen.vue'
import DemoMode from '@/components/home/DemoMode.vue'

const importAccountDeleteService = () => import('@/services/accountDelete')
import {success} from '@/message'

const authStore = useAuthStore()
const baseStore = useBaseStore()

const router = useRouter()
const route = useRoute()

useBodyClass('is-touch', isTouchDevice())
const keyboardShortcutsActive = computed(() => baseStore.keyboardShortcutsActive)

const {t} = useI18n({useScope: 'global'})

// setup account deletion verification
const accountDeletionConfirm = computed(() => route.query?.accountDeletionConfirm as (string | undefined))
watch(accountDeletionConfirm, async (accountDeletionConfirm) => {
	if (accountDeletionConfirm === undefined) {
		return
	}

	const AccountDeleteService = (await importAccountDeleteService()).default
	const accountDeletionService = new AccountDeleteService()
	await accountDeletionService.confirm(accountDeletionConfirm)
	success({message: t('user.deletion.confirmSuccess')})
	authStore.refreshUserInfo()
}, { immediate: true })

// setup email verification redirect
const userEmailConfirm = computed(() => route.query?.userEmailConfirm as (string | undefined))
watch(userEmailConfirm, (userEmailConfirm) => {
	if (userEmailConfirm === undefined) {
		return
	}

	localStorage.setItem('emailConfirmToken', userEmailConfirm)
	router.push({name: 'user.login'})
}, { immediate: true })

setLanguage(authStore.settings.language)
useColorScheme()
</script>

<style lang="scss" src="@/styles/global.scss" />
