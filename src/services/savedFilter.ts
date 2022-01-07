import AbstractService from '@/services/abstractService'
import SavedFilterModel from '@/models/savedFilter'
import {objectToCamelCase} from '@/helpers/case'

export default class SavedFilterService extends AbstractService {
	constructor() {
		super({
			get: '/filters/{id}',
			create: '/filters',
			update: '/filters/{id}',
			delete: '/filters/{id}',
		})
	}

	modelFactory(data) {
		return new SavedFilterModel(data)
	}

	processModel(model) {
		// Make filters from this.filters camelCase and set them to the model property:
		// That's easier than making the whole filter component configurable since that still needs to provide
		// the filter values in snake_sćase for url parameters.
		model.filters = objectToCamelCase(model.filters)

		// Make sure all filterValues are passes as strings. This is a requirement of the api.
		model.filters.filterValue = model.filters.filterValue.map(v => String(v))

		return model
	}

	beforeUpdate(model) {
		return this.processModel(model)
	}

	beforeCreate(model) {
		return this.processModel(model)
	}
}
