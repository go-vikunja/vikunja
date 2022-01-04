<template>
	<div>
		<message variant="danger" v-if="errorMessage !== ''">
			{{ errorMessage }}
		</message>
		<form @submit.prevent="submit" id="registerform">
			<div class="field">
				<label class="label" for="username">{{ $t('user.auth.username') }}</label>
				<div class="control">
					<input
						class="input"
						id="username"
						name="username"
						:placeholder="$t('user.auth.usernamePlaceholder')"
						required
						type="text"
						autocomplete="username"
						v-focus
						v-model="credentials.username"
						@keyup.enter="submit"
					/>
				</div>
			</div>
			<div class="field">
				<label class="label" for="email">{{ $t('user.auth.email') }}</label>
				<div class="control">
					<input
						class="input"
						id="email"
						name="email"
						:placeholder="$t('user.auth.emailPlaceholder')"
						required
						type="email"
						v-model="credentials.email"
						@keyup.enter="submit"
					/>
				</div>
			</div>
			<div class="field">
				<label class="label" for="password">{{ $t('user.auth.password') }}</label>
				<div class="control">
					<input
						class="input"
						id="password"
						name="password"
						:placeholder="$t('user.auth.passwordPlaceholder')"
						required
						type="password"
						autocomplete="new-password"
						v-model="credentials.password"
						@keyup.enter="submit"
					/>
				</div>
			</div>
			<div class="field">
				<label class="label" for="passwordValidation">{{ $t('user.auth.passwordRepeat') }}</label>
				<div class="control">
					<input
						class="input"
						id="passwordValidation"
						name="passwordValidation"
						:placeholder="$t('user.auth.passwordPlaceholder')"
						required
						type="password"
						autocomplete="new-password"
						v-model="passwordValidation"
						@keyup.enter="submit"
					/>
				</div>
			</div>

			<div class="field is-grouped">
				<div class="control">
					<x-button
						:loading="loading"
						id="register-submit"
						@click="submit"
						class="mr-2"
					>
						{{ $t('user.auth.register') }}
					</x-button>
					<x-button :to="{ name: 'user.login' }" variant="secondary">
						{{ $t('user.auth.login') }}
					</x-button>
				</div>
			</div>
		</form>
	</div>
</template>

<script setup>
import {ref, reactive, toRaw, computed, onBeforeMount} from 'vue'
import {useI18n} from 'vue-i18n'

import router from '@/router'
import {store} from '@/store'
import Message from '@/components/misc/message'

// FIXME: use the `beforeEnter` hook of vue-router
// Check if the user is already logged in, if so, redirect them to the homepage
onBeforeMount(() => {
	if (store.state.auth.authenticated) {
		router.push({name: 'home'})
	}
})

const {t} = useI18n()

const credentials = reactive({
	username: '',
	email: '',
	password: '',
})
const passwordValidation = ref('')

const loading = computed(() => store.state.loading)
const errorMessage = ref('')

async function submit() {
	errorMessage.value = ''

	if (credentials.password !== passwordValidation.value) {
		errorMessage.value = t('user.auth.passwordsDontMatch')
		return
	}


	try {
		await store.dispatch('auth/register', toRaw(credentials))
	} catch (e) {
		errorMessage.value = e.message
	}
}
</script>
