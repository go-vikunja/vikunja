import {ProjectFactory} from '../../factories/project'
import {TaskFactory} from '../../factories/task'
import {ProjectViewFactory} from "../../factories/project_view";

export function createDefaultViews(projectId) {
	ProjectViewFactory.truncate()
	const list = ProjectViewFactory.create(1, {
		id: 1,
		project_id: projectId,
		view_kind: 0,
	}, false)
	const gantt = ProjectViewFactory.create(1, {
		id: 2,
		project_id: projectId,
		view_kind: 1,
	}, false)
	const table = ProjectViewFactory.create(1, {
		id: 3,
		project_id: projectId,
		view_kind: 2,
	}, false)
	const kanban = ProjectViewFactory.create(1, {
		id: 4,
		project_id: projectId,
		view_kind: 3,
		bucket_configuration_mode: 1,
	}, false)

	return [
		list[0],
		gantt[0],
		table[0],
		kanban[0],
	]
}

export function createProjects() {
	const projects = ProjectFactory.create(1, {
		title: 'First Project'
	})
	TaskFactory.truncate()
	projects.views = createDefaultViews(projects[0].id)
	return projects
}

export function prepareProjects(setProjects = (...args: any[]) => {
}) {
	beforeEach(() => {
		const projects = createProjects()
		setProjects(projects)
	})
}
