import AbstractService from './abstractService'
import ProjectModel from '@/models/project'
import type {IProject} from '@/modelTypes/IProject'
import TaskService from './task'
import {colorFromHex} from '@/helpers/color/colorFromHex'

export default class ProjectService extends AbstractService<IProject> {
	constructor() {
		super({
			create: '/projects',
			get: '/projects/{id}',
			getAll: '/projects',
			update: '/projects/{id}',
			delete: '/projects/{id}',
		})
	}

	modelFactory(data) {
		return new ProjectModel(data)
	}

	beforeUpdate(model) {
		if(typeof model.tasks !== 'undefined') {
			const taskService = new TaskService()
			model.tasks = model.tasks.map(task => {
				return taskService.beforeUpdate(task)
			})
		}
		
		if(typeof model.hexColor !== 'undefined') {
			model.hexColor = colorFromHex(model.hexColor)
		}
		
		return model
	}

	beforeCreate(project) {
		project.hexColor = colorFromHex(project.hexColor)
		return project
	}

	async background(project: Pick<IProject, 'id' | 'backgroundInformation'>) {
		if (project.backgroundInformation === null) {
			return ''
		}

		const response = await this.http({
			url: `/projects/${project.id}/background`,
			method: 'GET',
			responseType: 'blob',
		})
		return window.URL.createObjectURL(new Blob([response.data]))
	}

	async removeBackground(project: IProject) {
		const cancel = this.setLoading()

		try {
			await this.http.delete(`/projects/${project.id}/background`)
			return {
				...project,
				backgroundInformation: null,
				backgroundBlurHash: '',
			}
		} finally {
			cancel()
		}
	}
}
