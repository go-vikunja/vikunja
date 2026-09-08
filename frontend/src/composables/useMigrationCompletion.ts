import {onScopeDispose, ref} from 'vue'

import type {MigrationStatus} from '@/services/migrator/abstractMigration'
import {parseDateOrNull} from '@/helpers/parseDateOrNull'
import {useProjectStore} from '@/stores/projects'

const POLL_INTERVAL = 3000
const POLL_DEADLINE = 20 * 60 * 1000
const MAX_CONSECUTIVE_FAILURES = 5

interface MigrationStatusSource {
	getStatus(): Promise<MigrationStatus>
}

// The migrate response only confirms the job started, not that it finished.

export function useMigrationCompletion(getSource: () => MigrationStatusSource) {
	const isFinished = ref(false)
	let timeout: ReturnType<typeof setTimeout> | undefined
	let generation = 0
	let deadline = 0
	let failures = 0

	function stop() {
		generation++
		clearTimeout(timeout)
		timeout = undefined
	}

	async function poll() {
		const myGeneration = generation

		try {
			const {finished_at} = await getSource().getStatus()
			if (myGeneration !== generation) {
				return
			}
			failures = 0

			if (parseDateOrNull(finished_at) !== null) {
				isFinished.value = true
				await useProjectStore().loadAllProjects()
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

	function start() {
		stop()
		isFinished.value = false
		failures = 0
		deadline = Date.now() + POLL_DEADLINE
		timeout = setTimeout(poll, POLL_INTERVAL)
	}

	onScopeDispose(stop)

	return {isFinished, start, stop}
}
