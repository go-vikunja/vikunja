import {Factory} from '../support/factory'

// Local "YYYY-MM-DD HH:MM:SS" (the format the DB fixtures use), not ISO-with-Z.
// start_time is filtered with datemath day windows that resolve to local time,
// and the comparison is lexical — a UTC-stamped value falls outside "today"
// near midnight.
function sqlDateTime(d: Date): string {
	const pad = (n: number) => String(n).padStart(2, '0')
	return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

export class TimeEntryFactory extends Factory {
	static table = 'time_entries'

	static factory() {
		const now = sqlDateTime(new Date())

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
