<template>
	<form
		class="credentials-form"
		novalidate
		@submit.prevent="submit"
	>
		<p>{{ $t('migrate.credentials.description', {name: migratorName}) }}</p>
		<Message
			v-if="error"
			variant="danger"
			class="mbe-4"
		>
			{{ error }}
		</Message>
		<FormField
			id="migration-url"
			v-model="credentials.url"
			v-focus
			:label="$t('migrate.credentials.url', {name: migratorName})"
			:placeholder="urlPlaceholder"
			:error="credentialsErrors.url"
			type="url"
			autocomplete="url"
		/>
		<fieldset class="field">
			<legend class="label">
				{{ $t('migrate.credentials.authMethod') }}
			</legend>
			<div class="control auth-method">
				<label class="radio">
					<input
						v-model="authMethod"
						type="radio"
						name="auth-method"
						value="token"
					>
					{{ $t('migrate.credentials.apiKey') }}
				</label>
				<label class="radio">
					<input
						v-model="authMethod"
						type="radio"
						name="auth-method"
						value="password"
					>
					{{ $t('migrate.credentials.authPassword') }}
				</label>
			</div>
		</fieldset>
		<template v-if="authMethod === 'token'">
			<FormField
				id="migration-token"
				v-model="credentials.token"
				:label="$t('migrate.credentials.apiKey')"
				:error="credentialsErrors.token"
				type="password"
				autocomplete="off"
				:aria-describedby="apiKeyHelp ? 'migration-token-help' : undefined"
			/>
			<p
				v-if="apiKeyHelp"
				id="migration-token-help"
				class="help"
			>
				{{ apiKeyHelp }}
			</p>
		</template>
		<template v-else>
			<FormField
				id="migration-username"
				v-model="credentials.username"
				:label="$t('user.auth.usernameEmail')"
				:error="credentialsErrors.username"
				type="text"
				autocomplete="off"
			/>
			<FormField
				id="migration-password"
				v-model="credentials.password"
				:label="$t('user.auth.password')"
				:error="credentialsErrors.password"
				type="password"
				autocomplete="off"
				:aria-describedby="passwordHelp ? 'migration-password-help' : undefined"
			/>
			<p
				v-if="passwordHelp"
				id="migration-password-help"
				class="help"
			>
				{{ passwordHelp }}
			</p>
		</template>
		<XButton
			type="submit"
			:loading="loading"
			:disabled="loading || undefined"
		>
			{{ $t('migrate.credentials.start') }}
		</XButton>
	</form>
</template>

<script setup lang="ts">
import {computed, reactive, ref} from 'vue'
import {useI18n} from 'vue-i18n'

import FormField from '@/components/input/FormField.vue'
import Message from '@/components/misc/Message.vue'

import type {MigrationConfig} from '@/services/migrator/abstractMigration'

const props = withDefaults(defineProps<{
	migratorName: string,
	loading: boolean,
	error: string,
	apiKeyHelp?: string,
	passwordHelp?: string,
}>(), {
	apiKeyHelp: '',
	passwordHelp: '',
})

const emit = defineEmits<{
	submit: [config: MigrationConfig],
	clearError: [],
}>()

const {t} = useI18n({useScope: 'global'})

const urlPlaceholder = computed(() => `https://${props.migratorName.toLowerCase()}.example.com`)

const authMethod = ref<'token' | 'password'>('token')
const credentials = reactive({url: '', token: '', username: '', password: ''})
const credentialsErrors = reactive<{
	url: string | null,
	token: string | null,
	username: string | null,
	password: string | null,
}>({url: null, token: null, username: null, password: null})

function submit() {
	emit('clearError')

	const url = credentials.url.trim()
	credentialsErrors.url = url === '' ? t('apiConfig.urlRequired') : null

	if (authMethod.value === 'token') {
		credentialsErrors.token = credentials.token.trim() === '' ? t('migrate.credentials.apiKeyRequired') : null
		credentialsErrors.username = null
		credentialsErrors.password = null
	} else {
		credentialsErrors.token = null
		credentialsErrors.username = credentials.username.trim() === '' ? t('user.auth.usernameRequired') : null
		credentialsErrors.password = credentials.password === '' ? t('user.auth.passwordRequired') : null
	}

	if (credentialsErrors.url || credentialsErrors.token || credentialsErrors.username || credentialsErrors.password) {
		return
	}

	emit('submit', authMethod.value === 'token'
		? {url, token: credentials.token.trim()}
		: {url, username: credentials.username.trim(), password: credentials.password})
}
</script>

<style lang="scss" scoped>
.credentials-form {
	max-inline-size: 500px;
}

fieldset {
	border: 0;
	margin: 0;
	padding: 0;
}

.auth-method {
	display: flex;
	flex-wrap: wrap;
	gap: 1rem;
}
</style>
