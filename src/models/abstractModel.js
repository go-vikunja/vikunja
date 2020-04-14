import {defaults, omitBy, isNil} from 'lodash'
import {objectToCamelCase} from '../helpers/case'

export default class AbstractModel {

	/**
	 * The abstract constructor takes an object and merges its data with the default data of this model.
	 * @param data
	 */
	constructor(data) {

		data = objectToCamelCase(data)

		// Put all data in our model while overriding those with a value of null or undefined with their defaults
		defaults(this, omitBy(data, isNil), this.defaults())
	}

	/**
	 * Default attributes that define the "empty" state.
	 * @return {{}}
	 */
	defaults() {
		return {}
	}
}