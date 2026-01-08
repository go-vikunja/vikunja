import AbstractService from './abstractService'
import WikiPageModel from '@/models/wikiPage'
import type {IWikiPage} from '@/modelTypes/IWikiPage'

export default class WikiPageService extends AbstractService<IWikiPage> {
	constructor() {
		super({
			create: '/projects/{projectId}/wiki',
			get: '/projects/{projectId}/wiki/{id}',
			getAll: '/projects/{projectId}/wiki',
			update: '/projects/{projectId}/wiki/{id}',
			delete: '/projects/{projectId}/wiki/{id}',
		})
	}

	modelFactory(data) {
		return new WikiPageModel(data)
	}

	processModel(model) {
		return {
			...model,
			// Ensure parent_id is sent correctly to backend
			parent_id: model.parentId,
		}
	}

	async move(projectId: number, pageId: number, newParentId: number | null) {
		const cancel = this.setLoading()
		
		try {
			const response = await this.http.post(`/projects/${projectId}/wiki/${pageId}/move`, {
				parent_id: newParentId,
			})
			return this.modelFactory(response.data)
		} finally {
			cancel()
		}
	}

	async reorder(projectId: number, pageId: number, position: number) {
		const cancel = this.setLoading()
		
		try {
			const response = await this.http.post(`/projects/${projectId}/wiki/${pageId}/reorder`, {
				position,
			})
			return this.modelFactory(response.data)
		} finally {
			cancel()
		}
	}

	async search(projectId: number, query: string): Promise<IWikiPage[]> {
		const cancel = this.setLoading()
		
		try {
			const response = await this.http.get(`/projects/${projectId}/wiki/search`, {
				params: {q: query},
			})
			return response.data.map(d => this.modelFactory(d))
		} finally {
			cancel()
		}
	}
}
