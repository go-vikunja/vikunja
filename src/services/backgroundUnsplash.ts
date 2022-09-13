import AbstractService from './abstractService'
import BackgroundImageModel from '../models/backgroundImage'
import ListModel from '@/models/list'
import type { IBackgroundImage } from '@/modelTypes/IBackgroundImage'

export default class BackgroundUnsplashService extends AbstractService<IBackgroundImage> {
	constructor() {
		super({
			getAll: '/backgrounds/unsplash/search',
			update: '/lists/{listId}/backgrounds/unsplash',
		})
	}

	modelFactory(data: Partial<IBackgroundImage>) {
		return new BackgroundImageModel(data)
	}

	modelUpdateFactory(data) {
		return new ListModel(data)
	}

	async thumb(model) {
		const response = await this.http({
			url: `/backgrounds/unsplash/images/${model.id}/thumb`,
			method: 'GET',
			responseType: 'blob',
		})
		return window.URL.createObjectURL(new Blob([response.data]))
	}
}