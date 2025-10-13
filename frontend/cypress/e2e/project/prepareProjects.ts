import {ProjectFactory} from '../../factories/project'
import {TaskFactory} from '../../factories/task'
import {ProjectViewFactory} from "../../factories/project_view";

export function createDefaultViews(projectId: number, startViewId = 1, truncate: boolean = true) {
	if (truncate) {
		ProjectViewFactory.truncate()
	}
	const list = ProjectViewFactory.create(1, {
		id: startViewId,
		project_id: projectId,
		view_kind: 0,
	}, false)
	const gantt = ProjectViewFactory.create(1, {
		id: startViewId + 1,
		project_id: projectId,
		view_kind: 1,
	}, false)
	const table = ProjectViewFactory.create(1, {
		id: startViewId + 2,
		project_id: projectId,
		view_kind: 2,
	}, false)
	const kanban = ProjectViewFactory.create(1, {
		id: startViewId + 3,
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

export function createProjects(count = 1) {
	const projects = ProjectFactory.create(count, {
		title: i => count === 1 ? 'First Project' : `Project ${i + 1}`,
	})
	TaskFactory.truncate()
	ProjectViewFactory.truncate()

	let viewIdCounter = 1
	for (let i = 0; i < projects.length; i++) {
		projects[i].views = createDefaultViews(projects[i].id, viewIdCounter, false)
		viewIdCounter += 4
	}

	return projects
}

export function prepareProjects(setProjects = (...args: any[]) => {
}) {
	beforeEach(() => {
		const projects = createProjects()
		setProjects(projects)
	})
}
