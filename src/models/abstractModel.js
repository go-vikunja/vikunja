import {objectToCamelCase} from '@/helpers/case'
import {omitBy, isNil} from '@/helpers/utils'

export default class AbstractModel {

	/**
	 * The max right the user has on this object, as returned by the x-max-right header from the api.
	 * @type {number|null}
	 */
	maxRight = null

	/**
	 * The abstract constructor takes an object and merges its data with the default data of this model.
	 * @param data
	 */
	constructor(data) {
		data = objectToCamelCase(data)

		// Put all data in our model while overriding those with a value of null or undefined with their defaults
		Object.assign(
			this,
			this.defaults(),
			omitBy(data, isNil),
		)
	}

	/**
	 * Default attributes that define the "empty" state.
	 * @return {{}}
	 */
	defaults() {
		return {}
	}
}