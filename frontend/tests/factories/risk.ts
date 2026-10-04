import {Factory} from '../support/factory'

// Local "YYYY-MM-DD HH:MM:SS", the format the DB fixtures use.
function sqlDateTime(d: Date): string {
	const pad = (n: number) => String(n).padStart(2, '0')
	return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

export class RiskFactory extends Factory {
	static table = 'risks'

	static factory() {
		const now = sqlDateTime(new Date())

		return {
			id: '{increment}',
			project_id: 1,
			title: (i: number) => `Risk ${i}`,
			description: '',
			category: '',
			probability: 3,
			impact: 3,
			owner_id: 0,
			mitigation: '',
			contingency: '',
			status: 'open',
			created_by_id: 1,
			created: now,
			updated: now,
		}
	}
}
