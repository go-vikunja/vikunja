import {reactive} from 'vue'

// Per row: a mutation's isPending/variables only track its latest call
export function usePendingRows() {
	const pending = reactive(new Set<number>())

	return {
		has: (id: number) => pending.has(id),
		async run(id: number, fn: () => Promise<unknown>) {
			if (pending.has(id)) {
				return
			}
			pending.add(id)
			try {
				await fn()
			} catch {
				// The mutation already reported it
			} finally {
				pending.delete(id)
			}
		},
	}
}
