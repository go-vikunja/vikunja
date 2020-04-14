import {camelCase} from 'camel-case'
import {snakeCase} from 'snake-case'

/**
 * Transforms field names to camel case.
 * @param object
 * @returns {*}
 */
export function objectToCamelCase(object) {
	let parsedObject = {}
	for (const m in object) {
		parsedObject[camelCase(m)] = object[m]
	}
	return parsedObject
}

/**
 * Transforms field names to snake case - used before making an api request.
 * @param object
 * @returns {*}
 */
export function objectToSnakeCase(object) {
	let parsedObject = {}
	for (const m in object) {
		parsedObject[snakeCase(m)] = object[m]
	}
	return parsedObject
}
