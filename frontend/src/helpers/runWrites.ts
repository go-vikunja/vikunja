// runWrites applies a write to each item. SQLite deadlocks on concurrent writes
// (read-then-write upgrade conflict), so callers pass concurrent=false to serialize.
export async function runWrites<T>(
	items: readonly T[],
	write: (item: T) => Promise<unknown>,
	concurrent: boolean,
): Promise<void> {
	if (concurrent) {
		await Promise.all(items.map(item => write(item)))
		return
	}
	for (const item of items) {
		await write(item)
	}
}
