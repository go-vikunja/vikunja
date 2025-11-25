import {ProjectFactory} from '../../factories/project'
import {TaskFactory} from '../../factories/task'
import {ProjectViewFactory} from '../../factories/project_view'

export async function createDefaultViews(projectId: number, startViewId = 1, truncate: boolean = true) {
	if (truncate) {
		ProjectViewFactory.truncate()
	}
	const list = await ProjectViewFactory.create(1, {
		id: startViewId,
		project_id: projectId,
		view_kind: 0,
	}, false)
	const gantt = await ProjectViewFactory.create(1, {
		id: startViewId + 1,
		project_id: projectId,
		view_kind: 1,
	}, false)
	const table = await ProjectViewFactory.create(1, {
		id: startViewId + 2,
		project_id: projectId,
		view_kind: 2,
	}, false)
	const kanban = await ProjectViewFactory.create(1, {
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

export async function createProjects(count: number = 1) {
	const projects = await ProjectFactory.create(count, {
		title: i => count === 1 ? 'First Project' : `Project ${i + 1}`,
	})

	TaskFactory.truncate()
	ProjectViewFactory.truncate()

	for (let i = 0; i < projects.length; i++) {
		const views = await createDefaultViews(projects[i].id, i * 4 + 1, false)
		projects[i].views = views
	}

	return projects
}
