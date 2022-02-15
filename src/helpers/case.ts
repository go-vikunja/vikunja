import {camelCase} from 'camel-case'
import {snakeCase} from 'snake-case'

/**
 * Transforms field names to camel case.
 * @param object
 * @returns {*}
 */
export function objectToCamelCase(object) {

	// When calling recursively, this can be called without being and object or array in which case we just return the value
	if (typeof object !== 'object') {
		return object
	}

	const parsedObject = {}
	for (const m in object) {
		parsedObject[camelCase(m)] = object[m]

		// Recursive processing
		// Prevent processing for some cases
		if (object[m] === null) {
			continue
		}

		// Call it again for arrays
		if (Array.isArray(object[m])) {
			parsedObject[camelCase(m)] = object[m].map(o => objectToCamelCase(o))
			// Because typeof [] === 'object' is true for arrays, we leave the loop here to prevent converting arrays to objects.
			continue
		}

		// Call it again for nested objects
		if (typeof object[m] === 'object') {
			parsedObject[camelCase(m)] = objectToCamelCase(object[m])
		}
	}
	return parsedObject
}

/**
 * Transforms field names to snake case - used before making an api request.
 * @param object
 * @returns {*}
 */
export function objectToSnakeCase(object) {

	// When calling recursively, this can be called without being and object or array in which case we just return the value
	if (typeof object !== 'object') {
		return object
	}

	const parsedObject = {}
	for (const m in object) {
		parsedObject[snakeCase(m)] = object[m]

		// Recursive processing
		// Prevent processing for some cases
		if (
			object[m] === null ||
			(object[m] instanceof Date)
		) {
			continue
		}

		// Call it again for arrays
		if (Array.isArray(object[m])) {
			parsedObject[snakeCase(m)] = object[m].map(o => objectToSnakeCase(o))
			// Because typeof [] === 'object' is true for arrays, we leave the loop here to prevent converting arrays to objects.
			continue
		}

		// Call it again for nested objects
		if (typeof object[m] === 'object') {
			parsedObject[snakeCase(m)] = objectToSnakeCase(object[m])
		}
	}

	return parsedObject
}
