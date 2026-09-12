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
						:loading="migrationFileService.loading"
						:disabled="migrationFileService.loading || undefined"
						@click="uploadInput?.click()"
					>
						{{ $t('migrate.upload') }}
					</XButton>
				</template>
				<MigrationCredentialsForm
					v-else-if="migrator.isCredentialsMigrator"
					:migrator-name="migrator.name"
					:loading="migrationService.loading"
					:error="migrationError"
					:api-key-help="apiKeyHelp"
					:password-help="passwordHelp"
					@submit="migrate"
					@clearError="migrationError = ''"
				/>
				<template v-else>
					<p>{{ $t('migrate.authorize', {name: migrator.name}) }}</p>
					<XButton
						:loading="migrationService.loading"
						:disabled="migrationService.loading || undefined"
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
				:variant="migrationFailureReason ? 'danger' : 'info'"
				role="status"
				aria-live="polite"
				tabindex="-1"
				class="mbe-4"
			>
				<template v-if="migrationFailureReason">
					{{ $t('migrate.migrationFailed', {service: migrator.name, reason: migrationFailureReason}) }}
				</template>
				<template v-else-if="migrationFinished">
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
export default {
	beforeRouteEnter(to) {
		if (MIGRATORS[to.params.service as string] === undefined) {
			return {name: 'not-found'}
		}
	},
}
</script>

<script setup lang="ts">
import {computed, nextTick, ref, shallowReactive, watch} from 'vue'
import {useI18n} from 'vue-i18n'

import Logo from '@/assets/logo.svg?component'
import Message from '@/components/misc/Message.vue'
import MigrationCredentialsForm from './MigrationCredentialsForm.vue'

import AbstractMigrationService, {type MigrationConfig} from '@/services/migrator/abstractMigration'
import AbstractMigrationFileService from '@/services/migrator/abstractMigrationFile'

import {formatDateLong} from '@/helpers/time/formatDate'
import {parseDateOrNull} from '@/helpers/parseDateOrNull'

import {MIGRATORS, type Migrator} from './migrators'
import {useTitle} from '@/composables/useTitle'
import {useMigrationCompletion} from '@/composables/useMigrationCompletion'
import {getErrorText} from '@/message'

const props = defineProps<{
	service: string,
	code?: string,
}>()

const PROGRESS_DOTS_COUNT = 8

const {t, te} = useI18n({useScope: 'global'})

const progressDotsCount = ref(PROGRESS_DOTS_COUNT)
const authUrl = ref('')
const isMigrating = ref(false)
const previousMigrationFinishedAt = ref<Date | null>(null)
const migrationRunning = ref(false)
const migratorAuthCode = ref('')
const migrationError = ref('')

const migrator = computed<Migrator>(() => MIGRATORS[props.service])

const apiKeyHelp = computed(() => {
	const key = `migrate.${migrator.value.id}.apiKeyHelp`
	return te(key) ? t(key) : ''
})
const passwordHelp = computed(() => {
	const key = `migrate.${migrator.value.id}.passwordHelp`
	return te(key) ? t(key) : ''
})

// eslint-disable-next-line vue/no-ref-object-reactivity-loss
const migrationService = shallowReactive(new AbstractMigrationService(migrator.value.id))
// eslint-disable-next-line vue/no-ref-object-reactivity-loss
const migrationFileService = shallowReactive(new AbstractMigrationFileService(migrator.value.id))

useTitle(() => t('migrate.titleService', {name: migrator.value.name}))

const statusSource = () => migrator.value.isFileMigrator ? migrationFileService : migrationService

const {
	isFinished: migrationFinished,
	errorMessage: migrationFailureReason,
	start: startPolling,
} = useMigrationCompletion(statusSource)

async function initMigration() {
	if (!migrator.value.isFileMigrator && !migrator.value.isCredentialsMigrator) {
		authUrl.value = await migrationService.getAuthUrl().then(({url}) => url)

		const TOKEN_HASH_PREFIX = '#token='
		migratorAuthCode.value = location.hash.startsWith(TOKEN_HASH_PREFIX)
			? location.hash.substring(TOKEN_HASH_PREFIX.length)
			: props.code as string

		if (!migratorAuthCode.value) {
			return
		}
	}

	const {started_at, finished_at} = await statusSource().getStatus()
	const finishedAt = parseDateOrNull(finished_at)

	if (parseDateOrNull(started_at) !== null && finishedAt === null) {
		migrationRunning.value = true
		startPolling()
		return
	}

	if (finishedAt !== null) {
		// A file migrator re-imports by uploading another file, so it needs the upload form, not a confirm prompt.
		if (!migrator.value.isFileMigrator) {
			previousMigrationFinishedAt.value = finishedAt
		}
		return
	}

	if (migrator.value.isFileMigrator || migrator.value.isCredentialsMigrator) {
		return
	}

	await migrate()
}

initMigration()

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

async function migrate(credentialsConfig?: MigrationConfig) {
	let migrationConfig: MigrationConfig | File = credentialsConfig ?? {code: migratorAuthCode.value}

	isMigrating.value = true
	previousMigrationFinishedAt.value = null
	migrationError.value = ''

	if (migrator.value.isFileMigrator) {
		if (uploadInput.value?.files?.length === 0) {
			return
		}
		migrationConfig = uploadInput.value?.files?.[0] as File
	}

	try {
		// The migrate response only means the import was queued, not finished.
		if (migrator.value.isFileMigrator) {
			await migrationFileService.migrate(migrationConfig as File)
		} else {
			await migrationService.migrate(migrationConfig as MigrationConfig)
		}
		migrationRunning.value = true
		startPolling()
	} catch (e) {
		migrationError.value = getErrorText(e)
	} finally {
		isMigrating.value = false
	}
}

// Credentials migrators have no config yet at this point, so reset the status to show the form again.
function confirmMigrateAgain() {
	if (!migrator.value.isCredentialsMigrator) {
		return migrate()
	}

	previousMigrationFinishedAt.value = null
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
