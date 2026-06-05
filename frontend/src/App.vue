<template>
	<Ready>
		<template v-if="isQuickAddMode && authStore.authUser">
			<QuickAddOverlay />
		</template>
		<template v-else-if="isQuickAddMode">
			<div class="quick-add-not-logged-in">
				<p>{{ $t('quickActions.notLoggedIn') }}</p>
			</div>
		</template>
		<template v-else>
			<a
				href="#main-content"
				class="skip-to-content"
			>
				{{ $t('misc.skipToContent') }}
			</a>
			<template v-if="showAuthLayout">
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
		</template>

		<KeyboardShortcuts v-if="keyboardShortcutsActive && !isQuickAddMode" />

		<Teleport to="body">
			<AddToHomeScreen v-if="!isQuickAddMode" />
			<UpdateNotification v-if="!isQuickAddMode" />
			<Notification />
			<DemoMode v-if="!isQuickAddMode" />
		</Teleport>
	</Ready>
</template>

<script lang="ts" setup>
import {computed, watch} from 'vue'
import {useRoute} from 'vue-router'
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

import {DEFAULT_LANGUAGE, setLanguage} from '@/i18n'

import {useAuthStore} from '@/stores/auth'
import {useBaseStore} from '@/stores/base'

import {useColorScheme} from '@/composables/useColorScheme'
import {useBodyClass} from '@/composables/useBodyClass'
import QuickAddOverlay from '@/components/quick-actions/QuickAddOverlay.vue'
import AddToHomeScreen from '@/components/home/AddToHomeScreen.vue'
import DemoMode from '@/components/home/DemoMode.vue'
import {AUTH_ROUTE_NAMES} from '@/constants/authRouteNames'
import {useQuickAddMode} from '@/composables/useQuickAddMode'

const importAccountDeleteService = () => import('@/services/accountDelete')
import {success} from '@/message'

const authStore = useAuthStore()
const baseStore = useBaseStore()

const {isQuickAddMode} = useQuickAddMode()

// Make the Electron frameless window transparent
if (isQuickAddMode) {
	document.documentElement.style.background = 'transparent'
	document.body.style.background = 'transparent'
}

const route = useRoute()

const showAuthLayout = computed(() => authStore.authUser && typeof route.name === 'string' && !AUTH_ROUTE_NAMES.has(route.name))

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

setLanguage(authStore.settings.language ?? DEFAULT_LANGUAGE)
useColorScheme()
</script>

<style src="@/styles/tailwind.css" />

<style lang="scss" src="@/styles/global.scss" />
