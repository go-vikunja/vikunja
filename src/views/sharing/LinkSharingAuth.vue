<template>
	<div>
		<message v-if="loading">
			{{ $t('sharing.authenticating') }}
		</message>
		<div v-if="authenticateWithPassword" class="box">
			<p class="pb-2">
				{{ $t('sharing.passwordRequired') }}
			</p>
			<div class="field">
				<div class="control">
					<input
						id="linkSharePassword"
						type="password"
						class="input"
						:placeholder="$t('user.auth.passwordPlaceholder')"
						v-model="password"
						v-focus
						@keyup.enter.prevent="authenticate()"
					/>
				</div>
			</div>

			<x-button @click="authenticate()" :loading="loading">
				{{ $t('user.auth.login') }}
			</x-button>

			<Message variant="danger" class="mt-4" v-if="errorMessage !== ''">
				{{ errorMessage }}
			</Message>
		</div>
	</div>
</template>

<script lang="ts" setup>
import {ref, computed} from 'vue'
import {useRoute, useRouter} from 'vue-router'
import {useI18n} from 'vue-i18n'
import {useTitle} from '@vueuse/core'

import Message from '@/components/misc/message.vue'
import {PROJECT_VIEWS, type ProjectView} from '@/types/ProjectView'
import {LINK_SHARE_HASH_PREFIX} from '@/constants/linkShareHash'

import {useBaseStore} from '@/stores/base'
import {useAuthStore} from '@/stores/auth'
import {useRedirectToLastVisited} from '@/composables/useRedirectToLastVisited'

const {t} = useI18n({useScope: 'global'})
useTitle(t('sharing.authenticating'))
const {getLastVisitedRoute} = useRedirectToLastVisited()

function useAuth() {
	const baseStore = useBaseStore()
	const authStore = useAuthStore()
	const route = useRoute()
	const router = useRouter()

	const loading = ref(false)
	const authenticateWithPassword = ref(false)
	const errorMessage = ref('')
	const password = ref('')

	const authLinkShare = computed(() => authStore.authLinkShare)

	async function authenticate() {
		authenticateWithPassword.value = false
		errorMessage.value = ''

		if (authLinkShare.value) {
			// FIXME: push to 'project.list' since authenticated?
			return
		}

		// TODO: no password

		loading.value = true

		try {
			const {project_id: projectId} = await authStore.linkShareAuth({
				hash: route.params.share,
				password: password.value,
			})
			const logoVisible = route.query.logoVisible
				? route.query.logoVisible === 'true'
				: true
			baseStore.setLogoVisible(logoVisible)

			const view = route.query.view && Object.values(PROJECT_VIEWS).includes(route.query.view as ProjectView)
				? route.query.view
				: 'list'

			const hash = LINK_SHARE_HASH_PREFIX + route.params.share

			const last = getLastVisitedRoute()
			if (last) {
				return router.push({
					...last,
					hash,
				})
			}

			return router.push({
				name: `project.${view}`,
				params: {projectId},
				hash,
			})
		} catch (e) {
			if (e?.response?.data?.code === 13001) {
				authenticateWithPassword.value = true
				return
			}

			// TODO: Put this logic in a global errorMessage handler method which checks all auth codes
			let err = t('sharing.error')
			if (e?.response?.data?.message) {
				err = e.response.data.message
			}
			if (e?.response?.data?.code === 13002) {
				err = t('sharing.invalidPassword')
			}
			errorMessage.value = err
		} finally {
			loading.value = false
		}
	}

	authenticate()

	return {
		loading,
		authenticateWithPassword,
		errorMessage,
		password,
		authenticate,
	}
}

const {
	loading,
	authenticateWithPassword,
	errorMessage,
	password,
	authenticate,
} = useAuth()
</script>
