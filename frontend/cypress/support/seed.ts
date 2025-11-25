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
export function seed(table, data = {}, truncate = true) {
	if (data === null) {
		data = []
	}

	cy.request({
		method: 'PATCH',
		url: `${Cypress.env('API_URL')}/test/${table}?truncate=${truncate ? 'true' : 'false'}`,
		headers: {
			'Authorization': Cypress.env('TEST_SECRET'),
		},
		body: data,
	})
}
