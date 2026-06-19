import {Factory} from '../support/factory'

export class TeamProjectFactory extends Factory {
	static table = 'team_projects'

	static factory() {
		const now = new Date()

		return {
			id: '{increment}',
			team_id: 1,
			project_id: 1,
			permission: 0,
			created: now.toISOString(),
			updated: now.toISOString(),
		}
	}
}
