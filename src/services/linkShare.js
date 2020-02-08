import AbstractService from './abstractService'
import LinkShareModel from '../models/linkShare'
import moment from 'moment'

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
        model.created = moment(model.created).toISOString()
        model.updated = moment(model.updated).toISOString()
        return model
    }

    modelFactory(data) {
        return new LinkShareModel(data)
    }
}