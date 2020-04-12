import AbstractService from './abstractService'
import LinkShareModel from '../models/linkShare'
import {formatISO} from 'date-fns'

export default class ListService extends AbstractService {
    constructor() {
        super({
            getAll: '/lists/{listId}/shares',
            get: '/lists/{listId}/shares/{id}',
            create: '/lists/{listId}/shares',
            delete: '/lists/{listId}/shares/{id}',
        })
    }

    processModel(model) {
        model.created = formatISO(model.created)
        model.updated = formatISO(model.updated)
        return model
    }

    modelFactory(data) {
        return new LinkShareModel(data)
    }
}