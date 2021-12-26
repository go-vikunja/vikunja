<template>
	<div>
		<message variant="danger" v-if="errorMessage !== ''" class="mb-4">
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
						@focusout="validateUsername"
					/>
				</div>
				<p class="help is-danger" v-if="!usernameValid">
					{{ $t('user.auth.usernameRequired') }}
				</p>
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
						@focusout="validateEmail"
					/>
				</div>
				<p class="help is-danger" v-if="!emailValid">
					{{ $t('user.auth.emailInvalid') }}
				</p>
			</div>
			<div class="field">
				<label class="label" for="password">{{ $t('user.auth.password') }}</label>
				<password @submit="submit" @update:modelValue="v => credentials.password = v" :validate-initially="validatePasswordInitially"/>
			</div>

			<x-button
				:loading="loading"
				id="register-submit"
				@click="submit"
				class="mr-2"
				:disabled="!everythingValid"
			>
				{{ $t('user.auth.createAccount') }}
			</x-button>
			<p class="mt-2">
				{{ $t('user.auth.alreadyHaveAnAccount') }}
				<router-link :to="{ name: 'user.login' }">
					{{ $t('user.auth.login') }}
				</router-link>
			</p>
		</form>
	</div>
</template>

<script setup>
import {useDebounceFn} from '@vueuse/core'
import {ref, reactive, toRaw, computed, onBeforeMount, watch} from 'vue'

import router from '@/router'
import {store} from '@/store'
import Message from '@/components/misc/message'
import {isEmail} from '@/helpers/isEmail'
import Password from '@/components/input/password'

// FIXME: use the `beforeEnter` hook of vue-router
// Check if the user is already logged in, if so, redirect them to the homepage
onBeforeMount(() => {
	if (store.state.auth.authenticated) {
		router.push({name: 'home'})
	}
})

const credentials = reactive({
	username: '',
	email: '',
	password: '',
})

const loading = computed(() => store.state.loading)
const errorMessage = ref('')
const validatePasswordInitially = ref(false)

const DEBOUNCE_TIME = 100

// debouncing to prevent error messages when clicking on the log in button
const emailValid = ref(true)
const validateEmail = useDebounceFn(() => {
	emailValid.value = isEmail(credentials.email)
}, DEBOUNCE_TIME)

const usernameValid = ref(true)
const validateUsername = useDebounceFn(() => {
	usernameValid.value = credentials.username !== ''
}, DEBOUNCE_TIME)

const everythingValid = computed(() => {
	return credentials.username !== '' &&
		credentials.email !== '' &&
		credentials.password !== '' &&
		emailValid.value &&
		usernameValid.value
})

async function submit() {
	errorMessage.value = ''
	validatePasswordInitially.value = true

	if (!everythingValid.value) {
		return
	}

	try {
		await store.dispatch('auth/register', toRaw(credentials))
	} catch (e) {
		errorMessage.value = e.message
	}
}
</script>
