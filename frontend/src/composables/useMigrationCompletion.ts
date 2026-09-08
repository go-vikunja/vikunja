import {onScopeDispose, ref} from 'vue'

import type {MigrationStatus} from '@/services/migrator/abstractMigration'
import {useProjectStore} from '@/stores/projects'

const POLL_INTERVAL = 3000

interface MigrationStatusSource {
	getStatus(): Promise<MigrationStatus>
}

/**
 * Imports run in the background, so the migrate response only confirms the job
 * started. Poll the migration status until it finishes, then pull in what it
 * created so the user doesn't have to reload.
 */
export function useMigrationCompletion(getSource: () => MigrationStatusSource) {
	const isFinished = ref(false)
	let timeout: ReturnType<typeof setTimeout> | undefined

	function stop() {
		clearTimeout(timeout)
		timeout = undefined
	}

	async function poll() {
		try {
			const {finished_at} = await getSource().getStatus()
			if (finished_at) {
				isFinished.value = true
				await useProjectStore().loadAllProjects()
				return
			}
		} catch {
			// The next tick retries, and the user is mailed either way.
		}
		timeout = setTimeout(poll, POLL_INTERVAL)
	}

	function start() {
		stop()
		isFinished.value = false
		timeout = setTimeout(poll, POLL_INTERVAL)
	}

	onScopeDispose(stop)

	return {isFinished, start, stop}
}
