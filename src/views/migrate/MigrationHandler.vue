<template>
	<div class="content">
		<h1>{{ $t('migrate.titleService', {name: migrator.name}) }}</h1>
		<p>{{ $t('migrate.descriptionDo') }}</p>

		<template v-if="message === '' && lastMigrationDate === null">
			<template v-if="isMigrating === false">
				<template v-if="migrator.isFileMigrator">
					<p>{{ $t('migrate.importUpload', {name: migrator.name}) }}</p>
					<input
						@change="migrate"
						class="is-hidden"
						ref="uploadInput"
						type="file"
					/>
					<x-button
						:loading="migrationFileService.loading"
						:disabled="migrationFileService.loading || undefined"
						@click="uploadInput?.click()"
					>
						{{ $t('migrate.upload') }}
					</x-button>
				</template>
				<template v-else>
					<p>{{ $t('migrate.authorize', {name: migrator.name}) }}</p>
					<x-button
						:loading="migrationService.loading"
						:disabled="migrationService.loading || undefined"
						:href="authUrl"
					>
						{{ $t('migrate.getStarted') }}
					</x-button>
				</template>
			</template>
			<div
				v-else
				class="migration-in-progress-container"
			>
				<div class="migration-in-progress">
					<img :alt="migrator.name" :src="migrator.icon" class="logo"/>
					<div class="progress-dots">
						<span v-for="i in progressDotsCount" :key="i"/>
					</div>
					<Logo class="logo"/>
				</div>
				<p>{{ $t('migrate.inProgress') }}</p>
			</div>
		</template>
		<div v-else-if="lastMigrationDate">
			<p>
				{{ $t('migrate.alreadyMigrated1', {name: migrator.name, date: formatDateLong(lastMigrationDate)}) }}<br/>
				{{ $t('migrate.alreadyMigrated2') }}
			</p>
			<div class="buttons">
				<x-button @click="migrate">{{ $t('migrate.confirm') }}</x-button>
				<x-button :to="{name: 'home'}" variant="tertiary" class="has-text-danger">{{ $t('misc.cancel') }}</x-button>
			</div>
		</div>
		<div v-else>
			<Message class="mb-4">
				{{ message }}
			</Message>
			<x-button :to="{name: 'home'}">{{ $t('misc.refresh') }}</x-button>
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
import {computed, ref, shallowReactive} from 'vue'
import {useI18n} from 'vue-i18n'

import Logo from '@/assets/logo.svg?component'
import Message from '@/components/misc/message.vue'

import AbstractMigrationService, { type MigrationConfig } from '@/services/migrator/abstractMigration'
import AbstractMigrationFileService from '@/services/migrator/abstractMigrationFile'

import {formatDateLong} from '@/helpers/time/formatDate'
import {parseDateOrNull} from '@/helpers/parseDateOrNull'

import {MIGRATORS} from './migrators'
import {useTitle} from '@/composables/useTitle'
import {useProjectStore} from '@/stores/projects'

const PROGRESS_DOTS_COUNT = 8

const props = defineProps<{
	service: string,
	code?: string,
}>()

const {t} = useI18n({useScope: 'global'})

const progressDotsCount = ref(PROGRESS_DOTS_COUNT)
const authUrl = ref('')
const isMigrating = ref(false)
const lastMigrationDate = ref<Date | null>(null)
const message = ref('')
const migratorAuthCode = ref('')

const migrator = computed(() => MIGRATORS[props.service])

const migrationService = shallowReactive(new AbstractMigrationService(migrator.value.id))
const migrationFileService = shallowReactive(new AbstractMigrationFileService(migrator.value.id))

useTitle(() => t('migrate.titleService', {name: migrator.value.name}))

async function initMigration() {
	if (migrator.value.isFileMigrator) {
		return
	}

	authUrl.value = await migrationService.getAuthUrl().then(({url}) => url)

	const TOKEN_HASH_PREFIX = '#token='
	migratorAuthCode.value = location.hash.startsWith(TOKEN_HASH_PREFIX)
		? location.hash.substring(TOKEN_HASH_PREFIX.length)
		: props.code as string

	if (!migratorAuthCode.value) {
		return
	}
	const {time} = await migrationService.getStatus()
	if (time) {
		lastMigrationDate.value = parseDateOrNull(time)

		if (lastMigrationDate.value) {
			return
		}
	}
	await migrate()
}

initMigration()

const uploadInput = ref<HTMLInputElement | null>(null)
async function migrate() {
	isMigrating.value = true
	lastMigrationDate.value = null
	message.value = ''

	let migrationConfig: MigrationConfig | File = {code: migratorAuthCode.value}

	if (migrator.value.isFileMigrator) {
		if (uploadInput.value?.files?.length === 0) {
			return
		}
		migrationConfig = uploadInput.value?.files?.[0] as File
	}

	try {
		const result = migrator.value.isFileMigrator
			? await migrationFileService.migrate(migrationConfig as File)
			: await migrationService.migrate(migrationConfig as MigrationConfig)
		message.value = result.message
		const projectStore = useProjectStore()
		return projectStore.loadProjects()
	} finally {
		isMigrating.value = false
	}
}
</script>

<style lang="scss" scoped>
.migration-in-progress-container {
	max-width: 400px;
	margin: 4rem auto 0;
	text-align: center;
}

.migration-in-progress {
	text-align: center;
	display: flex;
	max-width: 400px;
	justify-content: space-between;
	align-items: center;
	margin-bottom: 2rem;
}

.logo {
	display: block;
	max-height: 100px;
	max-width: 100px;
}

.progress-dots {
	height: 40px;
	width: 140px;
	overflow: visible;

	span {
		transition: all 500ms ease;
		background: var(--grey-500);
		height: 10px;
		width: 10px;
		display: inline-block;
		border-radius: 10px;
		animation: wave 2s ease infinite;
		margin-right: 5px;

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
</style>