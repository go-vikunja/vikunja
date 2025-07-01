import AbstractService from './abstractService'
import BackgroundImageModel from '../models/backgroundImage'
import ProjectModel from '@/models/project'
import type { IBackgroundImage } from '@/modelTypes/IBackgroundImage'
import type { IProject } from '@/modelTypes/IProject'

export default class BackgroundUnsplashService extends AbstractService<IBackgroundImage> {
	constructor() {
		super({
			getAll: '/backgrounds/unsplash/search',
			update: '/projects/{projectId}/backgrounds/unsplash',
		})
	}

	modelFactory(data: Partial<IBackgroundImage>) {
		return new BackgroundImageModel(data)
	}

	modelUpdateFactory(data: Partial<IProject>): IProject {
		return new ProjectModel(data)
	}

	async thumb(model: {id: string}) {
		const response = await this.http({
			url: `/backgrounds/unsplash/images/${model.id}/thumb`,
			method: 'GET',
			responseType: 'blob',
		})
		return window.URL.createObjectURL(new Blob([response.data]))
	}
}
