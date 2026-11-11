import {computed, ref, onScopeDispose, watch} from 'vue'
import {acceptHMRUpdate, defineStore} from 'pinia'
import {queryOptions, useMutation, useQuery} from '@tanstack/vue-query'
import {queryClient} from '@/client/queryClient'
import {captureClientRequestContext, isClientRequestContextCurrent} from '@/client/requestContext'
import {
	fetchMigrationStatus,
	migrationStatusQuery,
	migrationCompletedMutationOptions,
	type MigrationProvider,
} from '@/client/queries/migration'
import {parseDateOrNull} from '@/helpers/parseDateOrNull'

const POLL_INTERVAL = 3000
const POLL_DEADLINE = 20 * 60 * 1000
const POLL_RETRIES = 4

const GENERIC_FAILURE_KEY = 'migrate.failure.reported'

type MigrationErrorKind = 'reported' | 'interrupted' | 'credentials' | 'queue' | 'upload' | 'detail'

const FAILURE_KEYS: Partial<Record<MigrationErrorKind, string>> = {
	reported: GENERIC_FAILURE_KEY,
	interrupted: 'migrate.failure.interrupted',
	credentials: 'migrate.failure.credentials',
	queue: 'migrate.failure.queue',
	upload: 'migrate.failure.upload',
}

// Keep polling when the import view unmounts.
export const useMigrationStore = defineStore('migration', () => {
	const provider = ref<MigrationProvider>('csv')
	const startedAt = ref<number | null>(null)
	let request = captureClientRequestContext()

	function stop() {
		startedAt.value = null
	}

	const status = useQuery(computed(() => {
		const polledProvider = provider.value
		const runStartedAt = startedAt.value
		return queryOptions({
			...migrationStatusQuery(polledProvider),
			enabled: runStartedAt !== null,
			retry: POLL_RETRIES,
			retryDelay: POLL_INTERVAL,
			refetchIntervalInBackground: true,
			queryFn: ({signal}) => {
				// A run belongs to the session and server that started it; a later one must not adopt it.
				if (!isClientRequestContextCurrent(request)) {
					stop()
					throw new DOMException('Client request context changed', 'AbortError')
				}
				return fetchMigrationStatus(polledProvider, signal)
			},
			refetchInterval: ({state}) => {
				if (runStartedAt === null || Date.now() >= runStartedAt + POLL_DEADLINE) {
					return false
				}
				// Cache state older than the run is the previous import's and says nothing about this one.
				if (state.status === 'error' && state.errorUpdatedAt >= runStartedAt) {
					return false
				}
				if (state.dataUpdatedAt >= runStartedAt && parseDateOrNull(state.data?.finished_at) !== null) {
					return false
				}
				return POLL_INTERVAL
			},
		})
	}), queryClient)

	const completed = useMutation(migrationCompletedMutationOptions(), queryClient)
	const hasRunResult = computed(() => startedAt.value !== null && status.dataUpdatedAt.value >= startedAt.value)
	const isFinished = computed(() => hasRunResult.value && parseDateOrNull(status.data.value?.finished_at) !== null)
	const errorKind = computed(() => isFinished.value ? status.data.value?.error_kind ?? '' : '')
	const errorMessage = computed(() => isFinished.value ? status.data.value?.error_message ?? '' : '')
	const hasFailed = computed(() => errorKind.value !== '')
	const failureKey = computed(() => {
		if (errorKind.value === 'detail' && errorMessage.value !== '') {
			return 'migrate.migrationFailed'
		}
		// A newer api can report a kind this table does not have.
		return FAILURE_KEYS[errorKind.value as MigrationErrorKind] ?? GENERIC_FAILURE_KEY
	})

	watch(isFinished, finished => {
		if (finished && !hasFailed.value) {
			completed.mutate(undefined)
		}
	})

	function start(nextProvider: MigrationProvider) {
		provider.value = nextProvider
		request = captureClientRequestContext()
		startedAt.value = Date.now()
	}
	onScopeDispose(stop)
	return {isFinished, errorMessage, hasFailed, failureKey, start, stop}
})

if (import.meta.hot) {
	import.meta.hot.accept(acceptHMRUpdate(useMigrationStore, import.meta.hot))
}
