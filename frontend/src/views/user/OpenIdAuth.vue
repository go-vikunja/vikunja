<template>
	<div>
		<Message
			v-if="errorMessage"
			variant="danger"
		>
			{{ errorMessage }}
		</Message>
		<Message
			v-if="errorMessageFromQuery"
			variant="danger"
			class="mt-2"
		>
			{{ errorMessageFromQuery }}
		</Message>
		<Message v-if="loading">
			{{ $t('user.auth.authenticating') }}
		</Message>
	</div>
</template>


<script setup lang="ts">
import {ref, computed, onMounted} from 'vue'
import {useRoute} from 'vue-router'
import {useI18n} from 'vue-i18n'

import {getErrorText} from '@/message'
import Message from '@/components/misc/Message.vue'
import {useRedirectToLastVisited} from '@/composables/useRedirectToLastVisited'

import {useAuthStore} from '@/stores/auth'

defineOptions({name: 'Auth'})

const {t} = useI18n({useScope: 'global'})

const route = useRoute()
const {redirectIfSaved} = useRedirectToLastVisited()

const authStore = useAuthStore()

const loading = computed(() => authStore.isLoading)
const errorMessage = ref('')
const errorMessageFromQuery = computed(() => route.query.error)

async function authenticateWithCode() {
	// This component gets mounted twice: The first time when the actual auth request hits the frontend,
	// the second time after that auth request succeeded and the outer component "content-no-auth" isn't used
	// but instead the "content-auth" component is used. Because this component is just a route and thus
	// gets mounted as part of a <router-view/> which both the content-auth and content-no-auth components have,
	// this re-mounts the component, even if the user is already authenticated.
	// To make sure we only try to authenticate the user once, we set this "authenticating" lock in localStorage
	// which ensures only one auth request is done at a time. We don't simply check if the user is already
	// authenticated to not prevent the whole authentication if some user is already logged in.
	if (localStorage.getItem('authenticating')) {
		return
	}
	localStorage.setItem('authenticating', 'true')

	errorMessage.value = ''

	if (typeof route.query.error !== 'undefined') {
		localStorage.removeItem('authenticating')
		errorMessage.value = typeof route.query.message !== 'undefined'
			? route.query.message as string
			: t('user.auth.openIdGeneralError')
		return
	}

	const state = localStorage.getItem('state')
	if (typeof route.query.state === 'undefined' || route.query.state !== state) {
		localStorage.removeItem('authenticating')
		errorMessage.value = t('user.auth.openIdStateError')
		return
	}

	try {
		const provider = Array.isArray(route.params.provider) ? route.params.provider[0] : route.params.provider
		const code = Array.isArray(route.query.code) ? route.query.code[0] : route.query.code
		
		if (!provider || !code) {
			errorMessage.value = t('user.auth.openIdGeneralError')
			return
		}
		
		await authStore.openIdAuth({
			provider,
			code,
		})
		redirectIfSaved()
	} catch(e) {
		errorMessage.value = getErrorText(e)
	} finally {
		localStorage.removeItem('authenticating')
	}
}

onMounted(() => authenticateWithCode())
</script>
