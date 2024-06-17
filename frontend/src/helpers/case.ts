import {camelCase, snakeCase} from 'change-case'

/**
 * Transforms field names to camel case.
 */
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function objectToCamelCase(object: Record<string, any>) {

	// When calling recursively, this can be called without being and object or array in which case we just return the value
	if (typeof object !== 'object') {
		return object
	}

	// eslint-disable-next-line @typescript-eslint/no-explicit-any
	const parsedObject: Record<string, any> = {}
	for (const m in object) {
		parsedObject[camelCase(m)] = object[m]

		// Recursive processing
		// Prevent processing for some cases
		if (object[m] === null) {
			continue
		}

		// Call it again for arrays
		if (Array.isArray(object[m])) {
			// eslint-disable-next-line @typescript-eslint/no-explicit-any
			parsedObject[camelCase(m)] = object[m].map((o: Record<string, any>) => objectToCamelCase(o))
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
 */
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function objectToSnakeCase(object: Record<string, any>) {

	// When calling recursively, this can be called without being and object or array in which case we just return the value
	if (typeof object !== 'object') {
		return object
	}

	// eslint-disable-next-line @typescript-eslint/no-explicit-any
	const parsedObject: Record<string, any> = {}
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
			// eslint-disable-next-line @typescript-eslint/no-explicit-any
			parsedObject[snakeCase(m)] = object[m].map((o: Record<string, any>) => objectToSnakeCase(o))
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
