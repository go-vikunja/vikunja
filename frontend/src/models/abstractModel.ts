import {objectToCamelCase} from '@/helpers/case'
import {omitBy, isNil} from '@/helpers/utils'
import type {Permission} from '@/constants/permissions'
import type {IAbstract} from '@/modelTypes/IAbstract'

export default abstract class AbstractModel<Model extends IAbstract = IAbstract> implements IAbstract {


	/**
	 * The max permission the user has on this object, as returned by the x-max-permission header from the api.
	 */
	maxPermission: Permission | null = null
	
	/**
	* Takes an object and merges its data with the default data of this model.
	*/
	assignData(data: Partial<Model>) {
		data = objectToCamelCase(data)

		// Put all data in our model while overriding those with a value of null or undefined with their defaults
		Object.assign(this, omitBy(data, isNil))
	}
}
