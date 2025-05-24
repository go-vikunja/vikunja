import AbstractService from './abstractService'
import LinkShareModel from '@/models/linkShare'
import type {ILinkShare} from '@/modelTypes/ILinkShare'

export default class LinkShareService extends AbstractService<ILinkShare> {
	constructor() {
		super({
			getAll: '/projects/{projectId}/shares',
			get: '/projects/{projectId}/shares/{id}',
			create: '/projects/{projectId}/shares',
			delete: '/projects/{projectId}/shares/{id}',
		})
	}

	modelFactory(data) {
		return new LinkShareModel(data)
	}
}
