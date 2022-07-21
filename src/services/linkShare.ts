import AbstractService from './abstractService'
import LinkShareModel, { type ILinkShare } from '@/models/linkShare'
import {formatISO} from 'date-fns'

export default class LinkShareService extends AbstractService<ILinkShare> {
	constructor() {
		super({
			getAll: '/lists/{listId}/shares',
			get: '/lists/{listId}/shares/{id}',
			create: '/lists/{listId}/shares',
			delete: '/lists/{listId}/shares/{id}',
		})
	}

	processModel(model) {
		model.created = formatISO(new Date(model.created))
		model.updated = formatISO(new Date(model.updated))
		return model
	}

	modelFactory(data) {
		return new LinkShareModel(data)
	}
}