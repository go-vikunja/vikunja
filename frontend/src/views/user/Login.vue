<template>
	<div>
		<Message
			v-if="confirmedEmailSuccess"
			variant="success"
			text-align="center"
			class="mbe-4"
		>
			{{ $t('user.auth.confirmEmailSuccess') }}
		</Message>
		<Message
			v-if="errorMessage"
			variant="danger"
			class="mbe-4"
		>
			{{ errorMessage }}
		</Message>
		<form
			v-if="localAuthEnabled || ldapAuthEnabled"
			id="loginform"
			@submit.prevent="submit"
		>
			<div class="field">
				<label
					class="label"
					for="username"
				>{{ $t('user.auth.usernameEmail') }}</label>
				<div class="control">
					<input
						id="username"
						ref="usernameRef"
						v-focus
						class="input"
						name="username"
						:placeholder="$t('user.auth.usernamePlaceholder')"
						required
						type="text"
						autocomplete="username"
						tabindex="1"
						@keyup.enter="submit"
						@focusout="validateUsernameField()"
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
				<div class="label-with-link">
					<label
						class="label"
						for="password"
					>{{ $t('user.auth.password') }}</label>
					<RouterLink
						v-if="localAuthEnabled"
						:to="{ name: 'user.password-reset.request' }"
						class="reset-password-link"
						tabindex="6"
					>
						{{ $t('user.auth.forgotPassword') }}
					</RouterLink>
				</div>
				<Password
					v-model="password"
					tabindex="2"
					:validate-initially="validatePasswordInitially"
					:validate-min-length="false"
					@submit="submit"
				/>
			</div>
			<div
				v-if="needsTotpPasscode"
				class="field"
			>
				<label
					class="label"
					for="totpPasscode"
				>{{ $t('user.auth.totpTitle') }}</label>
				<div class="control">
					<input
						id="totpPasscode"
						ref="totpPasscode"
						v-focus
						autocomplete="one-time-code"
						class="input"
						:placeholder="$t('user.auth.totpPlaceholder')"
						required
						type="text"
						tabindex="3"
						inputmode="numeric"
						@keyup.enter="submit"
					>
				</div>
			</div>
			<div class="field">
				<label class="label">
					<input
						v-model="rememberMe"
						type="checkbox"
						class="mie-1"
					>
					{{ $t('user.auth.remember') }}
				</label>
			</div>

			<XButton
				:loading="isLoading"
				tabindex="4"
				@click="submit"
			>
				{{ $t('user.auth.login') }}
			</XButton>
			<p
				v-if="registrationEnabled"
				class="mbs-2"
			>
				{{ $t('user.auth.noAccountYet') }}
				<RouterLink
					:to="{ name: 'user.register' }"
					type="secondary"
					tabindex="5"
				>
					{{ $t('user.auth.createAccount') }}
				</RouterLink>
			</p>
		</form>

		<div
			v-if="hasOpenIdProviders"
			class="mbs-4"
		>
			<XButton
				v-for="(p, k) in openidConnect.providers"
				:key="k"
				variant="secondary"
				class="is-fullwidth mbs-2"
				@click="redirectToProvider(p)"
			>
				{{ $t('user.auth.loginWith', {provider: p.name}) }}
			</XButton>
		</div>
	</div>
</template>

<script setup lang="ts">
import {computed, onBeforeMount, ref} from 'vue'
import {useI18n} from 'vue-i18n'
import {useDebounceFn} from '@vueuse/core'

import Message from '@/components/misc/Message.vue'
import Password from '@/components/input/Password.vue'

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

const registrationEnabled = computed(() => configStore.auth.local.registrationEnabled)
const localAuthEnabled = computed(() => configStore.auth.local.enabled)
const ldapAuthEnabled = computed(() => configStore.auth.ldap.enabled)

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
	margin-block-end: .5rem;

	.label {
		margin-block-end: 0;
	}
}
</style>
