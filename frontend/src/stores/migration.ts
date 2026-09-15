import {computed, ref} from 'vue'
import {acceptHMRUpdate, defineStore} from 'pinia'

import type {MigrationErrorKind, MigrationStatus} from '@/services/migrator/abstractMigration'
import {parseDateOrNull} from '@/helpers/parseDateOrNull'
import {useProjectStore} from '@/stores/projects'

const POLL_INTERVAL = 3000
const POLL_DEADLINE = 20 * 60 * 1000
const MAX_CONSECUTIVE_FAILURES = 5

const GENERIC_FAILURE_KEY = 'migrate.failure.reported'

// A kind missing here - including one a newer api adds - falls back to the generic text
// instead of rendering an empty message.
const FAILURE_KEYS: Partial<Record<MigrationErrorKind, string>> = {
	reported: GENERIC_FAILURE_KEY,
	interrupted: 'migrate.failure.interrupted',
	credentials: 'migrate.failure.credentials',
	queue: 'migrate.failure.queue',
	upload: 'migrate.failure.upload',
}

export interface MigrationStatusSource {
	getStatus(): Promise<MigrationStatus>
}

// The migrate response only confirms the job started, not that it finished. Polling lives in a
// store so that leaving the migration view does not kill it - the import keeps creating projects.
export const useMigrationStore = defineStore('migration', () => {
	const isFinished = ref(false)
	const errorKind = ref<MigrationErrorKind>('')
	const errorMessage = ref('')

	let source: MigrationStatusSource | null = null
	let timeout: ReturnType<typeof setTimeout> | undefined
	let generation = 0
	let deadline = 0
	let failures = 0

	const hasFailed = computed(() => errorKind.value !== '')

	const failureKey = computed(() => {
		if (errorKind.value === 'detail' && errorMessage.value !== '') {
			return 'migrate.migrationFailed'
		}
		return FAILURE_KEYS[errorKind.value] ?? GENERIC_FAILURE_KEY
	})

	function stop() {
		generation++
		clearTimeout(timeout)
		timeout = undefined
	}

	async function applyStatus({finished_at, error_kind, error_message}: MigrationStatus) {
		if (parseDateOrNull(finished_at) === null) {
			return false
		}

		isFinished.value = true
		errorKind.value = error_kind ?? ''
		errorMessage.value = error_message ?? ''
		if (!hasFailed.value) {
			await useProjectStore().loadAllProjects()
		}
		return true
	}

	async function poll() {
		const myGeneration = generation

		try {
			const status = await source?.getStatus()
			if (myGeneration !== generation || status === undefined) {
				return
			}
			failures = 0

			if (await applyStatus(status)) {
				return
			}
		} catch {
			if (myGeneration !== generation) {
				return
			}
			failures++
		}

		// Giving up keeps isFinished false, so the "we will email you" copy stays true.
		if (failures >= MAX_CONSECUTIVE_FAILURES || Date.now() >= deadline) {
			return
		}

		timeout = setTimeout(poll, POLL_INTERVAL)
	}

	function start(statusSource: MigrationStatusSource) {
		stop()
		source = statusSource
		isFinished.value = false
		errorKind.value = ''
		errorMessage.value = ''
		failures = 0
		deadline = Date.now() + POLL_DEADLINE
		timeout = setTimeout(poll, POLL_INTERVAL)
	}

	return {
		isFinished,
		errorMessage,
		hasFailed,
		failureKey,
		start,
		stop,
	}
})

if (import.meta.hot) {
	import.meta.hot.accept(acceptHMRUpdate(useMigrationStore, import.meta.hot))
}
