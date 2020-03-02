import AbstractService from './abstractService'
import LinkShareModel from '../models/linkShare'
import {formatISO} from 'date-fns'

export default class ListService extends AbstractService {
    constructor() {
        super({
            getAll: '/lists/{listID}/shares',
            get: '/lists/{listID}/shares/{id}',
            create: '/lists/{listID}/shares',
            delete: '/lists/{listID}/shares/{id}',
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