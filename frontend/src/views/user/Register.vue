<template>
	<div v-if="configStore.auth.local.registrationEnabled">
		<Message
			v-if="errorMessage !== ''"
			variant="danger"
			class="mbe-4"
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
						@focusout="() => {validateUsername(); validateUsernameAfterFirst = true}"
						@keyup="handleUsernameKeyup"
					>
				</div>
				<p
					v-if="usernameError"
					class="help is-danger"
				>
					{{ usernameError }}
				</p>
			</div>
			<div class="field">
				<label
					class="label"
					for="email"
				>
					{{ $t('user.auth.email') }}
				</label>
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
						@focusout="() => {validateEmail(); validateEmailAfterFirst = true}"
						@keyup="handleEmailKeyup"
					>
				</div>
				<p
					v-if="emailError"
					class="help is-danger"
				>
					{{ emailError }}
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
				<p
					v-if="passwordError"
					class="help is-danger"
				>
					{{ passwordError }}
				</p>
			</div>

			<XButton
				id="register-submit"
				:loading="isLoading"
				class="mie-2"
				:disabled="!everythingValid"
				@click="submit"
			>
				{{ $t('user.auth.createAccount') }}
			</XButton>

			<Message
				v-if="configStore.demoModeEnabled"
				variant="warning"
				class="mbs-4"
			>
				{{ $t('demo.title') }}
				{{ $t('demo.accountWillBeDeleted') }}<br>
				<strong class="is-uppercase">{{ $t('demo.everythingWillBeDeleted') }}</strong>
			</Message>

			<p class="mbs-2">
				{{ $t('user.auth.alreadyHaveAnAccount') }}
				<RouterLink :to="{ name: 'user.login' }">
					{{ $t('user.auth.login') }}
				</RouterLink>
			</p>
		</form>
	</div>
	<Message
		v-else
		variant="warning"
	>
		{{ $t('user.auth.registrationDisabled') }}
	</Message>
</template>

<script setup lang="ts">
import {useDebounceFn} from '@vueuse/core'
import {computed, onBeforeMount, reactive, ref, toRaw} from 'vue'
import {useI18n} from 'vue-i18n'

import router from '@/router'
import Message from '@/components/misc/Message.vue'
import {isEmail} from '@/helpers/isEmail'
import Password from '@/components/input/Password.vue'
import {parseValidationErrors} from '@/helpers/parseValidationErrors'

import {useAuthStore} from '@/stores/auth'
import {useConfigStore} from '@/stores/config'
import {validatePassword} from '@/helpers/validatePasswort'

const {t} = useI18n()
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
const serverValidationErrors = ref<Record<string, string>>({})

const DEBOUNCE_TIME = 100

// debouncing to prevent error messages when clicking on the log in button
const emailValid = ref(true)
const validateEmailAfterFirst = ref(false)
const validateEmail = useDebounceFn(() => {
	emailValid.value = isEmail(credentials.email)
}, DEBOUNCE_TIME)

const usernameValid = ref<true | string>(true)
const validateUsernameAfterFirst = ref(false)
const validateUsername = useDebounceFn(() => {
	if (credentials.username === '') {
		usernameValid.value = t('user.auth.usernameRequired')
		return
	}

	if (credentials.username.indexOf(' ') !== -1) {
		usernameValid.value = t('user.auth.usernameMustNotContainSpace')
		return
	}

	if (credentials.username.indexOf('://') !== -1 || credentials.username.indexOf('.') !== -1) {
		usernameValid.value = t('user.auth.usernameMustNotLookLikeUrl')
		return
	}

	usernameValid.value = true
}, DEBOUNCE_TIME)

const everythingValid = computed(() => {
	return credentials.username !== '' &&
		credentials.email !== '' &&
		validatePassword(credentials.password) === true &&
		emailValid.value &&
		usernameValid.value
})

const usernameError = computed(() => {
	// Client-side validation takes priority
	if (usernameValid.value !== true) {
		return usernameValid.value
	}
	// Show server-side error if present
	return serverValidationErrors.value.username || null
})

const emailError = computed(() => {
	// Client-side validation takes priority
	if (!emailValid.value) {
		return t('user.auth.emailInvalid')
	}
	// Show server-side error if present
	return serverValidationErrors.value.email || null
})

const passwordError = computed(() => {
	// Show server-side error if present
	return serverValidationErrors.value.password || null
})

function handleUsernameKeyup() {
	if (validateUsernameAfterFirst.value) {
		validateUsername()
	}
	delete serverValidationErrors.value.username
}

function handleEmailKeyup() {
	if (validateEmailAfterFirst.value) {
		validateEmail()
	}
	delete serverValidationErrors.value.email
}

interface ApiValidationError {
	message?: string
	invalid_fields?: string[]
}

function isApiValidationError(error: unknown): error is ApiValidationError {
	return error !== null &&
		typeof error === 'object' &&
		('message' in error || 'invalid_fields' in error)
}

async function submit() {
	errorMessage.value = ''
	serverValidationErrors.value = {}
	validatePasswordInitially.value = true

	if (!everythingValid.value) {
		return
	}

	try {
		await authStore.register(toRaw(credentials))
	} catch (e: unknown) {
		// Parse field-specific validation errors
		if (isApiValidationError(e)) {
			const fieldErrors = parseValidationErrors(e)

			if (Object.keys(fieldErrors).length > 0) {
				// Apply field-level errors (computed properties will display them)
				serverValidationErrors.value = fieldErrors
			} else {
				// Fallback to general error message if no field errors
				errorMessage.value = e.message || 'Registration failed'
			}
		} else {
			errorMessage.value = 'Registration failed'
		}
	}
}
</script>
