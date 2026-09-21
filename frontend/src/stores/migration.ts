import {computed, ref, onScopeDispose} from 'vue'
import {acceptHMRUpdate, defineStore} from 'pinia'
import {useQuery} from '@tanstack/vue-query'
import {queryClient} from '@/client/queryClient'
import {captureClientRequestContext, isClientRequestContextCurrent} from '@/client/requestContext'
import {migrationStatusQuery, migrationCompletedMutationOptions, type MigrationProvider} from '@/client/queries/migration'
import {parseDateOrNull} from '@/helpers/parseDateOrNull'

const POLL_INTERVAL = 3000
const POLL_DEADLINE = 20 * 60 * 1000
const MAX_CONSECUTIVE_FAILURES = 5

const GENERIC_FAILURE_KEY = 'migrate.failure.reported'

const FAILURE_KEYS: Record<string, string> = {
	reported: GENERIC_FAILURE_KEY,
	interrupted: 'migrate.failure.interrupted',
	credentials: 'migrate.failure.credentials',
	queue: 'migrate.failure.queue',
	upload: 'migrate.failure.upload',
}

// Keep polling when the import view unmounts.
export const useMigrationStore = defineStore('migration', () => {
	const provider = ref<MigrationProvider>('csv')
	const accepted = ref(false)
	const status = useQuery(computed(() => ({...migrationStatusQuery(provider.value), enabled: false})), queryClient)
	const isFinished = computed(() => accepted.value && parseDateOrNull(status.data.value?.finished_at) !== null)
	const errorKind = computed(() => isFinished.value ? status.data.value?.error_kind ?? '' : '')
	const errorMessage = computed(() => isFinished.value ? status.data.value?.error_message ?? '' : '')
	const hasFailed = computed(() => errorKind.value !== '')
	const failureKey = computed(() => errorKind.value === 'detail' && errorMessage.value !== ''
		? 'migrate.migrationFailed' : FAILURE_KEYS[errorKind.value] ?? GENERIC_FAILURE_KEY)

	let timeout: ReturnType<typeof setTimeout> | undefined
	let generation = 0
	let deadline = 0
	let failures = 0
	let request = captureClientRequestContext()

	function stop() {
		generation++
		accepted.value = false
		clearTimeout(timeout)
		timeout = undefined
		void queryClient.cancelQueries({queryKey: migrationStatusQuery(provider.value).queryKey})
	}

	async function poll() {
		const myGeneration = generation
		if (!isClientRequestContextCurrent(request)) {
			stop()
			return
		}
		try {
			const result = await queryClient.fetchQuery(migrationStatusQuery(provider.value))
			if (myGeneration !== generation || !isClientRequestContextCurrent(request)) return
			accepted.value = true
			failures = 0
			if (parseDateOrNull(result?.finished_at) !== null) {
				if (!result?.error_kind) {
					await queryClient.getMutationCache().build(queryClient, migrationCompletedMutationOptions()).execute(undefined)
				}
				return
			}
		} catch {
			if (myGeneration !== generation || !isClientRequestContextCurrent(request)) return
			failures++
		}
		if (failures >= MAX_CONSECUTIVE_FAILURES || Date.now() >= deadline) return
		timeout = setTimeout(poll, POLL_INTERVAL)
	}

	function start(nextProvider: MigrationProvider) {
		stop()
		provider.value = nextProvider
		request = captureClientRequestContext()
		failures = 0
		deadline = Date.now() + POLL_DEADLINE
		timeout = setTimeout(poll, POLL_INTERVAL)
	}
	onScopeDispose(stop)
	return {isFinished, errorMessage, hasFailed, failureKey, start, stop}
})

if (import.meta.hot) {
	import.meta.hot.accept(acceptHMRUpdate(useMigrationStore, import.meta.hot))
}
