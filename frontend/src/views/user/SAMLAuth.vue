<template>
	<div>
		<Message
			v-if="errorMessage"
			variant="danger"
		>
			{{ errorMessage }}
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

import Message from '@/components/misc/Message.vue'
import {useRedirectToLastVisited} from '@/composables/useRedirectToLastVisited'
import {useAuthStore} from '@/stores/auth'
import {saveToken} from '@/helpers/auth'

defineOptions({name: 'SAMLAuth'})

const {t} = useI18n({useScope: 'global'})
const route = useRoute()
const {redirectIfSaved} = useRedirectToLastVisited()
const authStore = useAuthStore()

const loading = computed(() => authStore.isLoading)
const errorMessage = ref('')

async function authenticateWithToken() {
	if (localStorage.getItem('authenticating')) {
		return
	}
	localStorage.setItem('authenticating', 'true')

	errorMessage.value = ''

	if (typeof route.query.error !== 'undefined') {
		localStorage.removeItem('authenticating')
		errorMessage.value = t('user.auth.samlError')
		return
	}

	const token = route.query.token as string
	if (!token) {
		localStorage.removeItem('authenticating')
		errorMessage.value = t('user.auth.samlError')
		return
	}

	try {
		saveToken(token, true)
		await authStore.checkAuth()
		redirectIfSaved()
	} catch (_e) {
		errorMessage.value = t('user.auth.samlError')
	} finally {
		localStorage.removeItem('authenticating')
	}
}

onMounted(() => authenticateWithToken())
</script>
