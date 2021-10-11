import AbstractService from './abstractService'
import BackgroundImageModel from '../models/backgroundImage'
import ListModel from '../models/list'

export default class BackgroundUnsplashService extends AbstractService {
	constructor() {
		super({
			getAll: '/backgrounds/unsplash/search',
			update: '/lists/{listId}/backgrounds/unsplash',
		})
	}

	modelFactory(data) {
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