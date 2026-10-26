// Mirrors the server: a new timer starting at `start` stops the running one
// there. Null when the running timer stops at now instead.
export function backdatedStop(runningStart: Date, start: Date, now: Date): Date | null {
	return start > runningStart && start < now ? start : null
}
