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
import {useStore} from 'vuex'
import {useRoute, useRouter} from 'vue-router'
import {useI18n} from 'vue-i18n'
import {useTitle} from '@vueuse/core'

import Message from '@/components/misc/message.vue'

const {t} = useI18n()
useTitle(t('sharing.authenticating'))

function useAuth() {
	const store = useStore()
	const route = useRoute()
	const router = useRouter()

	const loading = ref(false)
	const authenticateWithPassword = ref(false)
	const errorMessage = ref('')
	const password = ref('')

	const authLinkShare = computed(() => store.getters['auth/authLinkShare'])

	async function authenticate() {
		authenticateWithPassword.value = false
		errorMessage.value = ''
	
		if (authLinkShare.value) {
			// FIXME: push to 'list.list' since authenticated?
			return
		}
	
		// TODO: no password
	
		loading.value = true

		try {
			const {list_id: listId} = await store.dispatch('auth/linkShareAuth', {
				hash: route.params.share,
				password: password.value,
			})
			router.push({name: 'list.list', params: {listId}})
		} catch (e: any) {
			if (e.response?.data?.code === 13001) {
				authenticateWithPassword.value = true
				return
			}

			// TODO: Put this logic in a global errorMessage handler method which checks all auth codes
			let err = t('sharing.error')
			if (e.response?.data?.message) {
				err = e.response.data.message
			}
			if (e.response?.data?.code === 13002) {
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
