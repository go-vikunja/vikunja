<template>
	<ready>
		<div :class="{'is-touch': isTouch}">
			<div :class="{'is-hidden': !online}">
				<template v-if="authUser">
					<top-navigation/>
					<content-auth/>
				</template>
				<content-link-share v-else-if="authLinkShare"/>
				<content-no-auth v-else/>
				<notification/>
			</div>

			<transition name="fade">
				<keyboard-shortcuts v-if="keyboardShortcutsActive"/>
			</transition>
		</div>
	</ready>
</template>

<script>
import {defineComponent} from 'vue'
import {mapState, mapGetters} from 'vuex'
import isTouchDevice from 'is-touch-device'

import Notification from './components/misc/notification'
import {KEYBOARD_SHORTCUTS_ACTIVE, ONLINE} from './store/mutation-types'
import KeyboardShortcuts from './components/misc/keyboard-shortcuts'
import TopNavigation from './components/home/topNavigation'
import ContentAuth from './components/home/contentAuth'
import ContentLinkShare from './components/home/contentLinkShare'
import ContentNoAuth from './components/home/contentNoAuth'
import {setLanguage} from './i18n'
import AccountDeleteService from '@/services/accountDelete'
import Ready from '@/components/misc/ready'
import {useColorScheme} from '@/composables/useColorScheme'

export default defineComponent({
	name: 'app',
	components: {
		ContentNoAuth,
		ContentLinkShare,
		ContentAuth,
		TopNavigation,
		KeyboardShortcuts,
		Notification,
		Ready,
	},
	beforeMount() {
		this.setupOnlineStatus()
		this.setupPasswortResetRedirect()
		this.setupEmailVerificationRedirect()
		this.setupAccountDeletionVerification()
	},
	beforeCreate() {
		setLanguage()
	},
	setup() {
		useColorScheme()
	},
	created() {
		// Make sure to always load the home route when running with electron
		if (this.$route.fullPath.endsWith('frontend/index.html')) {
			this.$router.push({name: 'home'})
		}
	},
	computed: {
		isTouch() {
			return isTouchDevice()
		},
		...mapState({
			online: ONLINE,
			keyboardShortcutsActive: KEYBOARD_SHORTCUTS_ACTIVE,
		}),
		...mapGetters('auth', [
			'authUser',
			'authLinkShare',
		]),
	},
	methods: {
		setupOnlineStatus() {
			this.$store.commit(ONLINE, navigator.onLine)
			window.addEventListener('online', () => this.$store.commit(ONLINE, navigator.onLine))
			window.addEventListener('offline', () => this.$store.commit(ONLINE, navigator.onLine))
		},
		setupPasswortResetRedirect() {
			if (typeof this.$route.query.userPasswordReset === 'undefined') {
				return
			}

			localStorage.setItem('passwordResetToken', this.$route.query.userPasswordReset)
			this.$router.push({name: 'user.password-reset.reset'})
		},
		setupEmailVerificationRedirect() {
			if (typeof this.$route.query.userEmailConfirm === 'undefined') {
				return
			}

			localStorage.setItem('emailConfirmToken', this.$route.query.userEmailConfirm)
			this.$router.push({name: 'user.login'})
		},
		async setupAccountDeletionVerification() {
			if (typeof this.$route.query.accountDeletionConfirm === 'undefined') {
				return
			}

			const accountDeletionService = new AccountDeleteService()
			await accountDeletionService.confirm(this.$route.query.accountDeletionConfirm)
			this.$message.success({message: this.$t('user.deletion.confirmSuccess')})
			this.$store.dispatch('auth/refreshUserInfo')
		},
	},
})
</script>

<style lang="scss">
@import '@/styles/global.scss';
</style>
