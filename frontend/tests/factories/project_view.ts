import {Factory} from '../support/factory'
import {faker} from '@faker-js/faker'

export class ProjectViewFactory extends Factory {
	static table = 'project_views'

	static factory() {
		const now = new Date()

		return {
			id: '{increment}',
			title: faker.lorem.words(3),
			project_id: '{increment}',
			view_kind: 0,
			created: now.toISOString(),
			updated: now.toISOString(),
		}
	}
}
