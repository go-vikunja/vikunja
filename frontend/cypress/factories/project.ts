import {Factory} from '../support/factory'
import {faker} from '@faker-js/faker'

export interface ProjectAttributes {
	id: number | '{increment}';
	title: string;
	owner_id: number;
	created: string;
	updated: string;
}

export class ProjectFactory extends Factory {
	static table = 'projects'

	static factory() {
		const now = new Date()

		return {
			id: '{increment}',
			title: faker.lorem.words(3),
			owner_id: 1,
			created: now.toISOString(),
			updated: now.toISOString(),
		}
	}
}
