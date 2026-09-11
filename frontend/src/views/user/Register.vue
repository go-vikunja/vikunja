<template>
	<template v-if="isInvite && inviteState.status !== 'ready'">
		<Message v-if="inviteState.status === 'loading'">
			{{ $t('misc.loading') }}
		</Message>
		<template v-else>
			<Message
				variant="danger"
				class="mbe-4"
			>
				{{ $t(inviteState.status === 'invalid' ? 'user.auth.inviteInvalid' : 'user.auth.inviteLoadFailed') }}
			</Message>
			<RouterLink
				:to="{name: 'user.login'}"
				class="inline-link"
			>
				{{ $t('user.auth.login') }}
			</RouterLink>
		</template>
	</template>
	<div v-else-if="isInvite || configStore.auth.local.registrationEnabled">
		<Message
			v-if="isInvite && inviteState.status === 'ready' && inviteState.link.teams?.length"
			class="mbe-4"
		>
			{{ $t('user.auth.inviteTeams', {teams: inviteState.link.teams.map(team => team.name).join(', ')}) }}
		</Message>
		<Message
			v-if="errorMessage !== ''"
			variant="danger"
			class="mbe-4"
		>
			{{ errorMessage }}
		</Message>
		<Message
			v-if="confirmEmailMessage !== ''"
			variant="success"
			class="mbe-4"
		>
			{{ confirmEmailMessage }}
		</Message>
		<form
			id="registerform"
			@submit.prevent="submit"
		>
			<FormField
				id="username"
				v-model="credentials.username"
				v-focus
				:label="$t('user.auth.username')"
				name="username"
				:placeholder="$t('user.auth.usernamePlaceholder')"
				required
				type="text"
				autocomplete="username"
				:error="usernameError"
				@keyup.enter="submit"
				@focusout="validateUsername(); validateUsernameAfterFirst = true"
				@keyup="handleUsernameKeyup"
			/>
			<FormField
				id="email"
				v-model="credentials.email"
				:label="$t('user.auth.email')"
				name="email"
				:placeholder="$t('user.auth.emailPlaceholder')"
				required
				type="email"
				:error="emailError"
				autocomplete="email"
				@keyup.enter="submit"
				@focusout="validateEmail(); validateEmailAfterFirst = true"
				@keyup="handleEmailKeyup"
			/>
			<div class="field">
				<label
					class="label"
					for="password"
				>{{ $t('user.auth.password') }}</label>
				<Password
					:validate-initially="validatePasswordInitially"
					autocomplete="new-password"
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
				<RouterLink
					:to="{ name: 'user.login' }"
					class="inline-link"
				>
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
import {computed, onBeforeMount, onUnmounted, reactive, ref, toRaw} from 'vue'
import {useI18n} from 'vue-i18n'

import router from '@/router'
import type {PublicInviteLink} from '@/client/generated'
import {hasInviteLink, checkInviteLink, onInviteLinkChange} from '@/client/inviteLink'
import Message from '@/components/misc/Message.vue'
import {isEmail} from '@/helpers/isEmail'
import Password from '@/components/input/Password.vue'
import FormField from '@/components/input/FormField.vue'
import {parseValidationErrors, type ValidationError} from '@/helpers/parseValidationErrors'

import {useRedirectToLastVisited} from '@/composables/useRedirectToLastVisited'
import {useAuthStore} from '@/stores/auth'
import {useConfigStore} from '@/stores/config'
import {validatePassword} from '@/helpers/validatePasswort'

const isInvite = ref(hasInviteLink())
type InviteState = {status: 'loading' | 'invalid' | 'error'} | {status: 'ready', link: PublicInviteLink}
const inviteState = ref<InviteState>({status: 'loading'})
let inviteCheck = 0
async function loadInvite() {
	isInvite.value = hasInviteLink()
	if (!isInvite.value) return
	const check = ++inviteCheck
	inviteState.value = {status: 'loading'}
	try {
		const {data: link} = await checkInviteLink()
		if (check !== inviteCheck) return
		inviteState.value = {status: 'ready', link}
	} catch (e) {
		if (check !== inviteCheck) return
		inviteState.value = {status: e instanceof Object && 'status' in e && e.status === 404 ? 'invalid' : 'error'}
	}
}
onBeforeMount(loadInvite)
const unsubscribeInvite = onInviteLinkChange(loadInvite)
onUnmounted(() => {
	inviteCheck++
	unsubscribeInvite()
})

const {t} = useI18n()
const authStore = useAuthStore()
const configStore = useConfigStore()
const {redirectIfSaved} = useRedirectToLastVisited()

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
const confirmEmailMessage = ref('')
const validatePasswordInitially = ref(false)
const serverValidationErrors = ref<Partial<Record<string, string>>>({})

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
		usernameValid.value === true
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

function isApiValidationError(error: unknown): error is ValidationError {
	return error !== null &&
		typeof error === 'object' &&
		('invalid_fields' in error || 'errors' in error)
}

async function submit() {
	errorMessage.value = ''
	serverValidationErrors.value = {}
	validatePasswordInitially.value = true

	if (!everythingValid.value || (isInvite.value && inviteState.value.status !== 'ready')) {
		return
	}

	try {
		if (isInvite.value) {
			await authStore.registerWithInvite(toRaw(credentials))
		} else {
			await authStore.register(toRaw(credentials))
		}
		redirectIfSaved()
	} catch (e: unknown) {
		if (isInvite.value && e instanceof Object && 'code' in e && e.code === 2005) {
			inviteState.value = {status: 'invalid'}
			return
		}
		// 1012 = email not confirmed: registration itself succeeded
		if (e instanceof Object && 'code' in e && e.code === 1012) {
			confirmEmailMessage.value = t('user.auth.registrationConfirmEmail')
			return
		}

		// Parse field-specific validation errors
		if (isApiValidationError(e)) {
			const fieldErrors = parseValidationErrors(e)

			if (Object.keys(fieldErrors).length > 0) {
				// Apply field-level errors (computed properties will display them)
				serverValidationErrors.value = fieldErrors
			} else {
				// Fallback to general error message if no field errors
				errorMessage.value = t('user.auth.registrationFailed')
			}
		} else if (e instanceof Object && 'message' in e && typeof e.message === 'string') {
			// Non-validation backend errors (e.g. duplicate username) - show their message
			errorMessage.value = e.message
		} else {
			errorMessage.value = t('user.auth.registrationFailed')
		}
	}
}
</script>

<style lang="scss" scoped>
// Underline links sitting inside body text so they're not distinguished by color alone
.inline-link {
	text-decoration: underline;
}
</style>
