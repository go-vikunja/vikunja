<template>
	<div>
		<message variant="success" text-align="center" class="mb-4" v-if="confirmedEmailSuccess">
			{{ $t('user.auth.confirmEmailSuccess') }}
		</message>
		<message variant="danger" v-if="errorMessage" class="mb-4">
			{{ errorMessage }}
		</message>
		<form @submit.prevent="submit" id="loginform" v-if="localAuthEnabled">
			<div class="field">
				<label class="label" for="username">{{ $t('user.auth.usernameEmail') }}</label>
				<div class="control">
					<input
						class="input" id="username"
						name="username"
						:placeholder="$t('user.auth.usernamePlaceholder')"
						ref="usernameRef"
						required
						type="text"
						autocomplete="username"
						v-focus
						@keyup.enter="submit"
						tabindex="1"
						@focusout="validateUsernameField()"
					/>
				</div>
				<p class="help is-danger" v-if="!usernameValid">
					{{ $t('user.auth.usernameRequired') }}
				</p>
			</div>
			<div class="field">
				<div class="label-with-link">
					<label class="label" for="password">{{ $t('user.auth.password') }}</label>
					<router-link
						:to="{ name: 'user.password-reset.request' }"
						class="reset-password-link"
						tabindex="6"
					>
						{{ $t('user.auth.forgotPassword') }}
					</router-link>
				</div>
				<Password tabindex="2" @submit="submit" v-model="password" :validate-initially="validatePasswordInitially"/>
			</div>
			<div class="field" v-if="needsTotpPasscode">
				<label class="label" for="totpPasscode">{{ $t('user.auth.totpTitle') }}</label>
				<div class="control">
					<input
						autocomplete="one-time-code"
						class="input"
						id="totpPasscode"
						:placeholder="$t('user.auth.totpPlaceholder')"
						ref="totpPasscode"
						required
						type="text"
						v-focus
						@keyup.enter="submit"
						tabindex="3"
						inputmode="numeric"
					/>
				</div>
			</div>
			<div class="field">
				<label class="label">
					<input type="checkbox" v-model="rememberMe" class="mr-1"/>
					{{ $t('user.auth.remember') }}
				</label>
			</div>

			<x-button
				@click="submit"
				:loading="isLoading"
				tabindex="4"
			>
				{{ $t('user.auth.login') }}
			</x-button>
			<p class="mt-2" v-if="registrationEnabled">
				{{ $t('user.auth.noAccountYet') }}
				<router-link
					:to="{ name: 'user.register' }"
					type="secondary"
					tabindex="5"
				>
					{{ $t('user.auth.createAccount') }}
				</router-link>
			</p>
		</form>

		<div
			v-if="hasOpenIdProviders"
			class="mt-4">
			<x-button
				v-for="(p, k) in openidConnect.providers"
				:key="k"
				@click="redirectToProvider(p)"
				variant="secondary"
				class="is-fullwidth mt-2"
			>
				{{ $t('user.auth.loginWith', {provider: p.name}) }}
			</x-button>
		</div>
	</div>
</template>

<script setup lang="ts">
import {computed, onBeforeMount, ref} from 'vue'
import {useI18n} from 'vue-i18n'
import {useDebounceFn} from '@vueuse/core'

import Message from '@/components/misc/message.vue'
import Password from '@/components/input/password.vue'

import {getErrorText} from '@/message'
import {redirectToProvider} from '@/helpers/redirectToProvider'
import {useRedirectToLastVisited} from '@/composables/useRedirectToLastVisited'

import {useAuthStore} from '@/stores/auth'
import {useConfigStore} from '@/stores/config'

import {useTitle} from '@/composables/useTitle'

const {t} = useI18n({useScope: 'global'})
useTitle(() => t('user.auth.login'))

const authStore = useAuthStore()
const configStore = useConfigStore()
const {redirectIfSaved} = useRedirectToLastVisited()

const registrationEnabled = computed(() => configStore.registrationEnabled)
const localAuthEnabled = computed(() => configStore.auth.local.enabled)

const openidConnect = computed(() => configStore.auth.openidConnect)
const hasOpenIdProviders = computed(() => openidConnect.value.enabled && openidConnect.value.providers?.length > 0)

const isLoading = computed(() => authStore.isLoading)

const confirmedEmailSuccess = ref(false)
const errorMessage = ref('')
const password = ref('')
const validatePasswordInitially = ref(false)
const rememberMe = ref(false)

const authenticated = computed(() => authStore.authenticated)

onBeforeMount(() => {
	authStore.verifyEmail().then((confirmed) => {
		confirmedEmailSuccess.value = confirmed
	}).catch((e: Error) => {
		errorMessage.value = e.message
	})

	// Check if the user is already logged in, if so, redirect them to the homepage
	if (authenticated.value) {
		redirectIfSaved()
	}
})

const usernameValid = ref(true)
const usernameRef = ref<HTMLInputElement | null>(null)
const validateUsernameField = useDebounceFn(() => {
	usernameValid.value = usernameRef.value?.value !== ''
}, 100)

const needsTotpPasscode = computed(() => authStore.needsTotpPasscode)
const totpPasscode = ref<HTMLInputElement | null>(null)

async function submit() {
	errorMessage.value = ''
	// Some browsers prevent Vue bindings from working with autofilled values.
	// To work around this, we're manually getting the values here instead of relying on vue bindings.
	// For more info, see https://kolaente.dev/vikunja/frontend/issues/78
	const credentials = {
		username: usernameRef.value?.value,
		password: password.value,
		longToken: rememberMe.value,
	}

	if (credentials.username === '' || credentials.password === '') {
		// Trigger the validation error messages
		validateUsernameField()
		validatePasswordInitially.value = true
		return
	}

	if (needsTotpPasscode.value) {
		credentials.totpPasscode = totpPasscode.value?.value
	}

	try {
		await authStore.login(credentials)
		authStore.setNeedsTotpPasscode(false)
	} catch (e) {
		if (e.response?.data.code === 1017 && !credentials.totpPasscode) {
			return
		}

		errorMessage.value = getErrorText(e)
	}
}
</script>

<style lang="scss" scoped>
.button {
	margin: 0 0.4rem 0 0;
}

.reset-password-link {
	display: inline-block;
}

.label-with-link {
	display: flex;
	justify-content: space-between;
	margin-bottom: .5rem;

	.label {
		margin-bottom: 0;
	}
}
</style>
