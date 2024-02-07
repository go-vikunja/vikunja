<template>
	<div>
		<Message
			v-if="errorMessage !== ''"
			variant="danger"
			class="mb-4"
		>
			{{ errorMessage }}
		</Message>
		<form
			id="registerform"
			@submit.prevent="submit"
		>
			<div class="field">
				<label
					class="label"
					for="username"
				>{{ $t('user.auth.username') }}</label>
				<div class="control">
					<input
						id="username"
						v-model="credentials.username"
						v-focus
						class="input"
						name="username"
						:placeholder="$t('user.auth.usernamePlaceholder')"
						required
						type="text"
						autocomplete="username"
						@keyup.enter="submit"
						@focusout="validateUsername"
					>
				</div>
				<p
					v-if="!usernameValid"
					class="help is-danger"
				>
					{{ $t('user.auth.usernameRequired') }}
				</p>
			</div>
			<div class="field">
				<label
					class="label"
					for="email"
				>{{ $t('user.auth.email') }}</label>
				<div class="control">
					<input
						id="email"
						v-model="credentials.email"
						class="input"
						name="email"
						:placeholder="$t('user.auth.emailPlaceholder')"
						required
						type="email"
						@keyup.enter="submit"
						@focusout="validateEmail"
					>
				</div>
				<p
					v-if="!emailValid"
					class="help is-danger"
				>
					{{ $t('user.auth.emailInvalid') }}
				</p>
			</div>
			<div class="field">
				<label
					class="label"
					for="password"
				>{{ $t('user.auth.password') }}</label>
				<Password
					:validate-initially="validatePasswordInitially"
					@submit="submit"
					@update:modelValue="v => credentials.password = v"
				/>
			</div>

			<x-button
				id="register-submit"
				:loading="isLoading"
				class="mr-2"
				:disabled="!everythingValid"
				@click="submit"
			>
				{{ $t('user.auth.createAccount') }}
			</x-button>
			
			<Message
				v-if="configStore.demoModeEnabled"
				variant="warning"
				class="mt-4"
			>
				{{ $t('demo.title') }}
				{{ $t('demo.accountWillBeDeleted') }}<br>
				<strong class="is-uppercase">{{ $t('demo.everythingWillBeDeleted') }}</strong>
			</Message>
			
			<p class="mt-2">
				{{ $t('user.auth.alreadyHaveAnAccount') }}
				<router-link :to="{ name: 'user.login' }">
					{{ $t('user.auth.login') }}
				</router-link>
			</p>
		</form>
	</div>
</template>

<script setup lang="ts">
import {useDebounceFn} from '@vueuse/core'
import {ref, reactive, toRaw, computed, onBeforeMount} from 'vue'

import router from '@/router'
import Message from '@/components/misc/message.vue'
import {isEmail} from '@/helpers/isEmail'
import Password from '@/components/input/password.vue'

import {useAuthStore} from '@/stores/auth'
import {useConfigStore} from '@/stores/config'

const authStore = useAuthStore()
const configStore = useConfigStore()

// FIXME: use the `beforeEnter` hook of vue-router
// Check if the user is already logged in, if so, redirect them to the homepage
onBeforeMount(() => {
	if (authStore.authenticated) {
		router.push({name: 'home'})
	}
})

const credentials = reactive({
	username: '',
	email: '',
	password: '',
})

const isLoading = computed(() => authStore.isLoading)
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
		await authStore.register(toRaw(credentials))
	} catch (e) {
		errorMessage.value = e?.message
	}
}
</script>
