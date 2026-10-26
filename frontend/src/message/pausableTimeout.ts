export type HoldReason = 'hover' | 'focus'

export function pausableTimeout(duration: number, onExpire: () => void) {
	const holds = new Set<HoldReason>()
	let remaining = duration
	let startedAt = Date.now()
	let timer: ReturnType<typeof setTimeout> | undefined = setTimeout(onExpire, remaining)
	let cancelled = false

	return {
		hold(reason: HoldReason) {
			if (holds.size === 0 && timer !== undefined) {
				clearTimeout(timer)
				timer = undefined
				remaining -= Date.now() - startedAt
			}
			holds.add(reason)
		},
		release(reason: HoldReason) {
			if (!holds.delete(reason) || holds.size > 0 || cancelled) {
				return
			}
			startedAt = Date.now()
			timer = setTimeout(onExpire, remaining)
		},
		cancel() {
			cancelled = true
			clearTimeout(timer)
			timer = undefined
		},
	}
}
