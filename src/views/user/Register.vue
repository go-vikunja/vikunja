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
				<div class="control is-relative">
					<input
						class="input"
						id="password"
						name="password"
						:placeholder="$t('user.auth.passwordPlaceholder')"
						required
						:type="passwordFieldType"
						autocomplete="new-password"
						v-model="credentials.password"
						@keyup.enter="submit"
						@focusout="validatePassword"
					/>
					<a
						@click="togglePasswordFieldType" 
						class="password-field-type-toggle"
						aria-label="passwordFieldType === 'password' ? $t('user.auth.showPassword') : $t('user.auth.hidePassword')"
						v-tooltip="passwordFieldType === 'password' ? $t('user.auth.showPassword') : $t('user.auth.hidePassword')">
						<icon :icon="passwordFieldType === 'password' ? 'eye' : 'eye-slash'"/>
					</a>
				</div>
				<p class="help is-danger" v-if="!passwordValid">
					{{ $t('user.auth.passwordRequired') }}
				</p>
			</div>

			<div class="field is-grouped">
				<div class="control">
					<x-button
						:loading="loading"
						id="register-submit"
						@click="submit"
						class="mr-2"
						:disabled="!everythingValid"
					>
						{{ $t('user.auth.createAccount') }}
					</x-button>
					<x-button :to="{ name: 'user.login' }" type="secondary">
						{{ $t('user.auth.login') }}
					</x-button>
				</div>
			</div>
		</form>
	</div>
</template>

<script setup>
import {useDebounceFn} from '@vueuse/core'
import {ref, reactive, toRaw, computed, onBeforeMount} from 'vue'

import router from '@/router'
import {store} from '@/store'
import Message from '@/components/misc/message'
import {isEmail} from '@/helpers/isEmail'

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

const passwordValid = ref(true)
const validatePassword = useDebounceFn(() => {
	passwordValid.value = credentials.password !== ''
}, DEBOUNCE_TIME)

const everythingValid = computed(() => {
	return credentials.username !== '' &&
		credentials.email !== '' &&
		credentials.password !== '' &&
		emailValid.value &&
		usernameValid.value &&
		passwordValid.value
})

const passwordFieldType = ref('password')
const togglePasswordFieldType = () => {
	passwordFieldType.value = passwordFieldType.value === 'password'
		? 'text'
		: 'password'
}

async function submit() {
	errorMessage.value = ''

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
