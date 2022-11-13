import {ProjectFactory} from '../../factories/project'
import {NamespaceFactory} from '../../factories/namespace'
import {TaskFactory} from '../../factories/task'

export function createProjects() {
	NamespaceFactory.create(1)
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