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
				@click.prevent="skipToMainContent"
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
				<RouterView v-if="showNoAuthRoute" />
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
import {useRoute, useRouter} from 'vue-router'
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
import {useTimeTrackingFavicon} from '@/composables/useTimeTrackingFavicon'
import {useBodyClass} from '@/composables/useBodyClass'
import QuickAddOverlay from '@/components/quick-actions/QuickAddOverlay.vue'
import AddToHomeScreen from '@/components/home/AddToHomeScreen.vue'
import DemoMode from '@/components/home/DemoMode.vue'
import {AUTH_ROUTE_NAMES} from '@/constants/authRouteNames'
import {useQuickAddMode} from '@/composables/useQuickAddMode'

import {useConfirmDeletionMutation} from '@/client/queries/accountDeletion'

const authStore = useAuthStore()
const baseStore = useBaseStore()

const {isQuickAddMode} = useQuickAddMode()

// Activating #main-content only scrolls natively, so focus has to be moved explicitly.
// Deferred a frame because a synchronous focus() did not stick in Safari.
function skipToMainContent() {
	requestAnimationFrame(() => document.getElementById('main-content')?.focus())
}

// Make the Electron frameless window transparent
if (isQuickAddMode) {
	document.documentElement.style.background = 'transparent'
	document.body.style.background = 'transparent'
}

const route = useRoute()

const showAuthLayout = computed(() => authStore.authUser && typeof route.name === 'string' && !AUTH_ROUTE_NAMES.has(route.name))

// The router guard bounces every other route to /login while logged out, so anything
// else reaching the logged-out shell means the auth state was cleared mid-navigation
// (logout, expired session) while the old route is still current. Mounting it there
// would run app components against a null `authStore.info`.
const showNoAuthRoute = computed(() => typeof route.name === 'string' && AUTH_ROUTE_NAMES.has(route.name))

useBodyClass('is-touch', isTouchDevice())
const keyboardShortcutsActive = computed(() => baseStore.keyboardShortcutsActive)

const router = useRouter()
const confirmDeletion = useConfirmDeletionMutation()

// Without a session the token would burn on an unauthenticated request, so keep it in the URL until login lands.
const accountDeletionToken = computed(() => {
	const token = route.query.accountDeletionConfirm
	return authStore.authUser && typeof token === 'string' ? token : ''
})

watch(accountDeletionToken, async token => {
	if (!token) return
	// The URL reaches Sentry replays and lastVisited, so drop the single-use token before spending it.
	const query = {...route.query}
	delete query.accountDeletionConfirm
	// A failed strip must not skip spending the token.
	await router.replace({path: route.path, query, hash: route.hash}).catch(() => {})
	try {
		await confirmDeletion.mutateAsync(token)
	} catch {
		return
	} finally {
		confirmDeletion.reset()
	}
}, {immediate: true})

setLanguage(authStore.settings.language ?? DEFAULT_LANGUAGE)
useColorScheme()
useTimeTrackingFavicon()
</script>

<style src="@/styles/tailwind.css" />

<style lang="scss" src="@/styles/global.scss" />
