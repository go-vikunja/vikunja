import {ProjectFactory} from '../../factories/project'
import {TaskFactory} from '../../factories/task'

export function createProjects() {
	const projects = ProjectFactory.create(1, {
		title: 'First Project'
	})
	TaskFactory.truncate()
	return projects
}

export function prepareProjects(setProjects = (...args: any[]) => {}) {
	beforeEach(() => {
		const projects = createProjects()
		setProjects(projects)
	})
}