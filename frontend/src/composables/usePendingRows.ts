import {reactive} from 'vue'

type RowMutation = {mutateAsync: (id: number) => Promise<unknown>}

// Per row: a mutation's isPending/variables only track its latest call
export function usePendingRows() {
	const pending = reactive(new Set<number>())

	return {
		has: (id: number) => pending.has(id),
		async run(mutation: RowMutation, id: number): Promise<boolean> {
			if (pending.has(id)) {
				return false
			}
			pending.add(id)
			try {
				await mutation.mutateAsync(id)
				return true
			} catch {
				// The mutation's onError already toasted it
				return false
			} finally {
				pending.delete(id)
			}
		},
	}
}
