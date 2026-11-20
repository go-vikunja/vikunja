import {Factory} from '../support/factory'

export class TimeEntryFactory extends Factory {
	static table = 'time_entries'

	static factory() {
		// UTC, like every other factory: the seeder reads a zone-less string as UTC
		// (pkg/db/dump.go), so a local wall-clock stamp would be stored off by the
		// runner's UTC offset and fall outside the "Today" filter after local midnight.
		// The current instant is inside "today" in every zone.
		const now = new Date().toISOString()

		return {
			id: '{increment}',
			user_id: 1,
			task_id: 0,
			project_id: 0,
			// Completed by default (end set), within today so the default filter shows it.
			start_time: now,
			end_time: now,
			comment: '',
			created: now,
			updated: now,
		}
	}
}
