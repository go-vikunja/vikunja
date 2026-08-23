import AbstractService from './abstractService'
import BackgroundImageModel from '../models/backgroundImage'
import ProjectModel from '@/models/project'
import {apiV2Url} from '@/helpers/fetcher'
import type { IBackgroundImage } from '@/modelTypes/IBackgroundImage'
import type { IProject } from '@/modelTypes/IProject'

// Reuses the shared /backgrounds/unsplash/search and thumb endpoints (generic,
// exist on both v1 and v2) — only the "set" endpoint is card-specific and v2-only.
export default class CardBackgroundUnsplashService extends AbstractService<IBackgroundImage> {
	constructor() {
		super({
			getAll: '/backgrounds/unsplash/search',
		})
	}

	modelFactory(data: Partial<IBackgroundImage>) {
		return new BackgroundImageModel(data)
	}

	modelUpdateFactory(data) {
		return new ProjectModel(data)
	}

	async thumb(model) {
		const response = await this.http({
			url: `/backgrounds/unsplash/images/${model.id}/thumb`,
			method: 'GET',
			responseType: 'blob',
		})
		return window.URL.createObjectURL(new Blob([response.data]))
	}

	// v2 rejects unknown properties, so only {id} is sent — projectId is
	// path-only, matching how EmailUpdateService (@/services/emailUpdate) does it.
	async update(model: {id: string, projectId: IProject['id']}) {
		const cancel = this.setLoading()
		try {
			const response = await this.http.put(
				apiV2Url(`projects/${model.projectId}/card-backgrounds/unsplash`),
				{id: model.id},
			)
			return this.modelUpdateFactory(response.data)
		} finally {
			cancel()
		}
	}
}
