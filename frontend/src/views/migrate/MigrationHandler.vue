<template>
	<div class="content">
		<h1>{{ $t('migrate.titleService', {name: migrator.name}) }}</h1>
		<p>{{ $t('migrate.descriptionDo') }}</p>

		<template v-if="!migrationRunning && previousMigrationFinishedAt === null">
			<!-- the credentials form stays mounted while migrating so its input survives an error -->
			<template v-if="isMigrating === false || migrator.isCredentialsMigrator">
				<template v-if="migrator.isFileMigrator">
					<p>{{ $t('migrate.importUpload', {name: migrator.name}) }}</p>
					<Message
						v-if="migrationError"
						variant="danger"
						class="mbe-4"
					>
						{{ migrationError }}
					</Message>
					<input
						ref="uploadInput"
						class="is-hidden"
						type="file"
						@change="migrate()"
					>
					<XButton
						:loading="isMigrating"
						:disabled="isMigrating || undefined"
						@click="uploadInput?.click()"
					>
						{{ $t('migrate.upload') }}
					</XButton>
				</template>
				<MigrationCredentialsForm
					v-else-if="migrator.isCredentialsMigrator"
					:migrator-name="migrator.name"
					:loading="isMigrating"
					:error="migrationError"
					:api-key-help="apiKeyHelp"
					:password-help="passwordHelp"
					@submit="migrate"
					@clearError="migrationError = ''"
				/>
				<template v-else>
					<p>{{ $t('migrate.authorize', {name: migrator.name}) }}</p>
					<Message
						v-if="migrationError"
						variant="danger"
						class="mbe-4"
					>
						{{ migrationError }}
					</Message>
					<XButton
						:loading="isBusy"
						:disabled="isBusy || undefined"
						:href="authUrl"
						:open-external-in-new-tab="false"
					>
						{{ $t('migrate.getStarted') }}
					</XButton>
				</template>
			</template>
			<div
				v-else
				class="migration-in-progress-container"
			>
				<div class="migration-in-progress">
					<img
						:alt="migrator.name"
						:src="migrator.icon"
						class="logo"
					>
					<div class="progress-dots">
						<span
							v-for="i in progressDotsCount"
							:key="i"
						/>
					</div>
					<Logo class="logo" />
				</div>
				<p>{{ $t('migrate.inProgress') }}</p>
			</div>
		</template>
		<div v-else-if="previousMigrationFinishedAt">
			<p>
				{{
					$t('migrate.alreadyMigrated1', {name: migrator.name, date: formatDateLong(previousMigrationFinishedAt)})
				}}<br>
				{{ $t('migrate.alreadyMigrated2') }}
			</p>
			<div class="migration-buttons">
				<XButton @click="confirmMigrateAgain">
					{{ $t('migrate.confirm') }}
				</XButton>
				<XButton
					:to="{name: 'home'}"
					variant="tertiary"
					class="has-text-danger"
				>
					{{ $t('misc.cancel') }}
				</XButton>
			</div>
		</div>
		<div v-else>
			<Message
				ref="resultMessage"
				:variant="migrationStore.hasFailed ? 'danger' : 'info'"
				role="status"
				aria-live="polite"
				tabindex="-1"
				class="mbe-4"
			>
				<template v-if="migrationStore.hasFailed">
					{{ $t(migrationStore.failureKey, {service: migrator.name, reason: migrationStore.errorMessage}) }}
				</template>
				<template v-else-if="migrationStore.isFinished">
					{{ $t('migrate.migrationFinished', {service: migrator.name}) }}
				</template>
				<template v-else>
					{{ $t('migrate.migrationStartedWillReciveEmail', {service: migrator.name}) }}
				</template>
			</Message>

			<XButton :to="{name: 'home'}">
				{{ $t('home.goToOverview') }}
			</XButton>
		</div>
	</div>
</template>

<script lang="ts">
function isKnownMigrator(service: unknown) {
	return Object.prototype.hasOwnProperty.call(MIGRATORS, service as string)
}

export default {
	beforeRouteEnter(to) {
		if (!isKnownMigrator(to.params.service)) {
			return {name: 'not-found'}
		}
	},
	beforeRouteUpdate(to) {
		if (!isKnownMigrator(to.params.service)) {
			return {name: 'not-found'}
		}
	},
}
</script>

<script setup lang="ts">
import {computed, nextTick, ref, watch} from 'vue'
import {useI18n} from 'vue-i18n'

import Logo from '@/assets/logo.svg?component'
import Message from '@/components/misc/Message.vue'
import MigrationCredentialsForm, {type PlankaCredentials} from './MigrationCredentialsForm.vue'

import {useQuery} from '@tanstack/vue-query'
import {migrationStatusQuery, useMigrationAuthMutation, useStartMigrationMutation} from '@/client/queries/migration'

import {isRequestContextAbort} from '@/client/requestContext'

import {formatDateLong} from '@/helpers/time/formatDate'
import {parseDateOrNull} from '@/helpers/parseDateOrNull'

import {MIGRATORS, type Migrator} from './migrators'
import {useTitle} from '@/composables/useTitle'
import {useMigrationStore} from '@/stores/migration'
import {getErrorText} from '@/message'

const props = defineProps<{
	service: string,
	code?: string,
}>()

const PROGRESS_DOTS_COUNT = 8

const {t, te} = useI18n({useScope: 'global'})

const progressDotsCount = ref(PROGRESS_DOTS_COUNT)
const authUrl = ref('')
const auth = useMigrationAuthMutation()
const startMigration = useStartMigrationMutation()
const isMigrating = computed(() => startMigration.isPending.value)
const isBusy = computed(() => isMigrating.value || auth.isPending.value)
const confirmedAgain = ref(false)
const startedHere = ref(false)

const migratorAuthCode = ref('')
const migrationError = ref('')

const migrator = computed<Migrator>(() => MIGRATORS[props.service as keyof typeof MIGRATORS])

const apiKeyHelp = computed(() => {
	const key = `migrate.${migrator.value.id}.apiKeyHelp`
	return te(key) ? t(key) : ''
})
const passwordHelp = computed(() => {
	const key = `migrate.${migrator.value.id}.passwordHelp`
	return te(key) ? t(key) : ''
})

const status = useQuery(computed(() => ({
	...migrationStatusQuery(migrator.value.id),
	enabled: false,
	meta: {handlesError: true},
})))
const migrationRunning = computed(() => {
	if (startedHere.value) return true
	return parseDateOrNull(status.data.value?.started_at) !== null
		&& parseDateOrNull(status.data.value?.finished_at) === null
})
const previousMigrationFinishedAt = computed(() => {
	if (confirmedAgain.value || startedHere.value || migrator.value.isFileMigrator) return null
	return parseDateOrNull(status.data.value?.finished_at)
})
useTitle(() => t('migrate.titleService', {name: migrator.value.name}))

const migrationStore = useMigrationStore()

async function initMigration() {
	try {
		const provider = migrator.value.id
		const current = await status.refetch()
		if (provider !== migrator.value.id) return
		if (current.isError) throw current.error
		if (migrationRunning.value) {
			startedHere.value = true
			migrationStore.start(provider)
			return
		}
		if (provider !== 'todoist' && provider !== 'trello' && provider !== 'microsoft-todo') return
		const authResult = await auth.mutateAsync(provider)
		if (provider !== migrator.value.id) return
		authUrl.value = authResult?.url ?? ''
		const prefix = '#token='
		migratorAuthCode.value = location.hash.startsWith(prefix) ? location.hash.substring(prefix.length) : props.code ?? ''
		if (migratorAuthCode.value && previousMigrationFinishedAt.value === null) await migrate()
	} catch (cause) {
		if (!isRequestContextAbort(cause)) migrationError.value = getErrorText(cause)
	}
}
watch(() => props.service, () => {
	confirmedAgain.value = false
	startedHere.value = false
	authUrl.value = ''
	migrationError.value = ''
	void initMigration()
}, {immediate: true})

const uploadInput = ref<HTMLInputElement | null>(null)
const resultMessage = ref<InstanceType<typeof Message> | null>(null)

// the triggering button unmounts when the result message appears, so move focus there
watch(migrationRunning, async (running) => {
	if (!running) {
		return
	}
	await nextTick()
	resultMessage.value?.$el?.focus()
})

async function migrate(credentialsConfig?: PlankaCredentials) {
	const provider = migrator.value.id
	confirmedAgain.value = true
	migrationError.value = ''
	try {
		if (provider === 'ticktick' || provider === 'wekan' || provider === 'vikunja-file') {
			const file = uploadInput.value?.files?.[0]
			if (!file) return
			await startMigration.mutateAsync({kind: 'file', provider, file})
		} else if (provider === 'planka') {
			if (!credentialsConfig) return
			await startMigration.mutateAsync({kind: 'credentials', provider, body: credentialsConfig})
		} else if (provider === 'todoist' || provider === 'trello' || provider === 'microsoft-todo') {
			if (!migratorAuthCode.value) return
			await startMigration.mutateAsync({kind: 'oauth', provider, body: {code: migratorAuthCode.value}})
		} else return
		migrationStore.start(provider)
		if (provider === migrator.value.id) startedHere.value = true
	} catch (cause) {
		if (!isRequestContextAbort(cause)) migrationError.value = getErrorText(cause)
	} finally {
		// Evicts the plaintext password from the mutation cache.
		startMigration.reset()
	}
}
function confirmMigrateAgain() {
	confirmedAgain.value = true
	if (!migrator.value.isCredentialsMigrator) return migrate()
}
</script>

<style lang="scss" scoped>
.migration-in-progress-container {
	max-inline-size: 400px;
	margin: 4rem auto 0;
	text-align: center;
}

.migration-in-progress {
	text-align: center;
	display: flex;
	max-inline-size: 400px;
	justify-content: space-between;
	align-items: center;
	margin-block-end: 2rem;
}

.logo {
	display: block;
	max-block-size: 100px;
	max-inline-size: 100px;
}

.progress-dots {
	block-size: 40px;
	inline-size: 140px;
	overflow: visible;

	span {
		transition: all 500ms ease;
		background: var(--grey-500);
		block-size: 10px;
		inline-size: 10px;
		display: inline-block;
		border-radius: 10px;
		animation: wave 2s ease infinite;
		margin-inline-end: 5px;

		&:nth-child(1) {
			animation-delay: 0;
		}

		&:nth-child(2) {
			animation-delay: 100ms;
		}

		&:nth-child(3) {
			animation-delay: 200ms;
		}

		&:nth-child(4) {
			animation-delay: 300ms;
		}

		&:nth-child(5) {
			animation-delay: 400ms;
		}

		&:nth-child(6) {
			animation-delay: 500ms;
		}

		&:nth-child(7) {
			animation-delay: 600ms;
		}

		&:nth-child(8) {
			animation-delay: 700ms;
		}
	}
}

@keyframes wave {
	0%, 40%, 100% {
		transform: translate(0, 0);
		background-color: var(--primary);
	}
	10% {
		transform: translate(0, -15px);
		background-color: var(--primary-dark);
	}
}

@media (prefers-reduced-motion: reduce) {
	@keyframes wave {
		10% {
			transform: translate(0, 0);
			background-color: var(--primary);
		}
	}
}

.migration-buttons {
	display: flex;
	flex-wrap: wrap;
	align-items: center;
	justify-content: flex-start;
	gap: 0.5rem;
}
</style>
