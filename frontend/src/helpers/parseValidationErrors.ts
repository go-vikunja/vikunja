interface ValidationError {
	message?: string
	code?: number
	invalid_fields?: string[]
}

/**
 * Parses validation errors from API responses into a field-to-error map.
 * Extracts field names and messages from the invalid_fields array.
 *
 * @param error - The error object from API response
 * @returns Object mapping field names to error messages
 *
 * @example
 * // Returns: { email: "email is not a valid email address" }
 * parseValidationErrors({
 *   message: 'invalid data',
 *   invalid_fields: ['email: email is not a valid email address']
 * })
 */
export function parseValidationErrors(error: ValidationError | null | undefined): Record<string, string> {
	if (!error || !error.invalid_fields || error.invalid_fields.length === 0) {
		return {}
	}

	const fieldErrors: Record<string, string> = {}

	for (const fieldError of error.invalid_fields) {
		// Split on first colon to separate field name from message
		const colonIndex = fieldError.indexOf(':')
		if (colonIndex === -1) {
			// No field prefix, can't map to a specific field, skip it
			continue
		}

		// Extract field name and error message
		const fieldName = fieldError.substring(0, colonIndex).trim()
		const errorMessage = fieldError.substring(colonIndex + 1).trim()

		fieldErrors[fieldName] = errorMessage
	}

	return fieldErrors
}
