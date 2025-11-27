import type {APIRequestContext} from '@playwright/test'

/**
 * Seeds a db table with data. If a data object is provided as the second argument, it will load the fixtures
 * file for the table and merge the data from it with the passed data. This allows you to override specific
 * fields of the fixtures without having to redeclare the whole fixture.
 *
 * Passing null as the second argument empties the table.
 *
 * @param table
 * @param data
 */
export async function seed(apiContext: APIRequestContext, table: string, data: unknown = {}, truncate = true) {
	if (data === null) {
		data = []
	}

	const apiUrl = process.env.API_URL || 'http://localhost:3456/api/v1'
	const testSecret = process.env.TEST_SECRET || 'averyLongSecretToSe33dtheDB'

	await apiContext.patch(`${apiUrl}/test/${table}?truncate=${truncate ? 'true' : 'false'}`, {
		headers: {
			'Authorization': testSecret,
		},
		data: data,
	})
}
