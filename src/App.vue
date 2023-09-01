<template>
	<ready>
		<template v-if="authUser">
			<TheNavigation/>
			<content-auth/>
		</template>
		<content-link-share v-else-if="authLinkShare"/>
		<no-auth-wrapper v-else>
			<router-view/>
		</no-auth-wrapper>
		
		<keyboard-shortcuts v-if="keyboardShortcutsActive"/>
		
		<Teleport to="body">
			<AddToHomeScreen/>
			<UpdateNotification/>
			<Notification/>
			<DemoMode/>
		</Teleport>
	</ready>
</template>

<script lang="ts" setup>
import {computed, watch} from 'vue'
import {useRoute, useRouter} from 'vue-router'
import {useI18n} from 'vue-i18n'
import isTouchDevice from 'is-touch-device'

import Notification from '@/components/misc/notification.vue'
import UpdateNotification from '@/components/home/UpdateNotification.vue'
import KeyboardShortcuts from '@/components/misc/keyboard-shortcuts/index.vue'

import TheNavigation from '@/components/home/TheNavigation.vue'
import ContentAuth from '@/components/home/contentAuth.vue'
import ContentLinkShare from '@/components/home/contentLinkShare.vue'
import NoAuthWrapper from '@/components/misc/no-auth-wrapper.vue'
import Ready from '@/components/misc/ready.vue'

import {setLanguage} from '@/i18n'
import AccountDeleteService from '@/services/accountDelete'
import {success} from '@/message'

import {useAuthStore} from '@/stores/auth'
import {useBaseStore} from '@/stores/base'

import {useColorScheme} from '@/composables/useColorScheme'
import {useBodyClass} from '@/composables/useBodyClass'
import AddToHomeScreen from '@/components/home/AddToHomeScreen.vue'
import DemoMode from '@/components/home/DemoMode.vue'

const baseStore = useBaseStore()
const authStore = useAuthStore()
const router = useRouter()
const route = useRoute()

useBodyClass('is-touch', isTouchDevice())
const keyboardShortcutsActive = computed(() => baseStore.keyboardShortcutsActive)

const authUser = computed(() => authStore.authUser)
const authLinkShare = computed(() => authStore.authLinkShare)

const {t} = useI18n({useScope: 'global'})

// setup account deletion verification
const accountDeletionConfirm = computed(() => route.query?.accountDeletionConfirm as (string | undefined))
watch(accountDeletionConfirm, async (accountDeletionConfirm) => {
	if (accountDeletionConfirm === undefined) {
		return
	}

	const accountDeletionService = new AccountDeleteService()
	await accountDeletionService.confirm(accountDeletionConfirm)
	success({message: t('user.deletion.confirmSuccess')})
	authStore.refreshUserInfo()
}, { immediate: true })

// setup password reset redirect
const userPasswordReset = computed(() => route.query?.userPasswordReset as (string | undefined))
watch(userPasswordReset, (userPasswordReset) => {
	if (userPasswordReset === undefined) {
		return
	}

	localStorage.setItem('passwordResetToken', userPasswordReset)
	router.push({name: 'user.password-reset.reset'})
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

<style lang="scss">
@import '@/styles/global.scss';
</style>
