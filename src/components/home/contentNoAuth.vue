<template>
	<no-auth-wrapper>
		<router-view/>
	</no-auth-wrapper>
</template>

<script>
import {saveLastVisited} from '@/helpers/saveLastVisited'
import NoAuthWrapper from '@/components/misc/no-auth-wrapper'

export default {
	name: 'contentNoAuth',
	components: {NoAuthWrapper},
	computed: {
		routeName() {
			return this.$route.name
		},
	},
	watch: {
		routeName: {
			handler(routeName) {
				if (!routeName) return
				this.redirectToHome()
			},
			immediate: true,
		},
	},
	methods: {
		redirectToHome() {
			// Check if the user is already logged in and redirect them to the home page if not
			if (
				this.$route.name !== 'user.login' &&
				this.$route.name !== 'user.password-reset.request' &&
				this.$route.name !== 'user.password-reset.reset' &&
				this.$route.name !== 'user.register' &&
				this.$route.name !== 'link-share.auth' &&
				this.$route.name !== 'openid.auth' &&
				localStorage.getItem('passwordResetToken') === null &&
				localStorage.getItem('emailConfirmToken') === null
			) {
				saveLastVisited(this.$route.name, this.$route.params)
				this.$router.push({name: 'user.login'})
			}
		},
	},
}
</script>
