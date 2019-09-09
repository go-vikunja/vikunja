import AbstractService from './abstractService'
import LinkShareModel from '../models/linkShare'

export default class ListService extends AbstractService {
    constructor() {
        super({
            getAll: '/lists/{listID}/shares',
            get: '/lists/{listID}/shares/{id}',
            create: '/lists/{listID}/shares',
            delete: '/lists/{listID}/shares/{id}',
        })
    }

    modelFactory(data) {
        return new LinkShareModel(data)
    }
}