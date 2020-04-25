import AbstractService from './abstractService'
import BucketModel from "../models/bucket";

export default class BucketService extends AbstractService {
	constructor() {
		super({
			getAll: '/lists/{listId}/buckets',
			create: '/lists/{listId}/buckets',
			update: '/lists/{listId}/buckets/{id}',
			delete: '/lists/{listId}/buckets/{id}',
		})
	}

	modelFactory(data) {
		return new BucketModel(data)
	}
}