<template>
	<no-auth-wrapper>
		<router-view/>
	</no-auth-wrapper>
</template>

<script lang="ts" setup>
import {watchEffect} from 'vue'
import {useRoute, useRouter} from 'vue-router'

import NoAuthWrapper from '@/components/misc/no-auth-wrapper'

import {saveLastVisited} from '@/helpers/saveLastVisited'

const route = useRoute()

watchEffect(() => {
	if (!route.name) return
	redirectToHome()
})

const router = useRouter()
function redirectToHome() {
	// Check if the user is already logged in and redirect them to the home page if not
	if (
		![
			'user.login',
			'user.password-reset.request',
			'user.password-reset.reset',
			'user.register',
			'link-share.auth',
			'openid.auth',
		].includes(route.name) &&
		localStorage.getItem('passwordResetToken') === null &&
		localStorage.getItem('emailConfirmToken') === null
	) {
		saveLastVisited(route.name, route.params)
		router.push({name: 'user.login'})
	}
}
</script>
