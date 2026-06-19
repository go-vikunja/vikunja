import {faker} from '@faker-js/faker'
import {Factory} from '../support/factory'

export class TaskFactory extends Factory {
	static table = 'tasks'

	static factory() {
		const now = new Date()

		return {
			id: '{increment}',
			title: faker.lorem.words(3),
			done: false,
			project_id: 1,
			created_by_id: 1,
			index: '{increment}',
			created: now.toISOString(),
			updated: now.toISOString(),
		}
	}

	// Mirror numeric `id` overrides onto `index` so sequential single-row
	// creates don't collide on UNIQUE(project_id, index). Matches the
	// id == index convention used by raw seedTasks helpers.
	static async create(count = 1, override: Record<string, unknown> = {}, truncate = true) {
		if (
			typeof override.id === 'number' &&
			!('index' in override)
		) {
			override = {...override, index: override.id}
		}
		return super.create(count, override, truncate)
	}
}
