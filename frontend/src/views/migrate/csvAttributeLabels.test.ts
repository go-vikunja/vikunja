import {readFileSync} from 'node:fs'
import {resolve} from 'node:path'

import {describe, expect, it} from 'vitest'

import {CSV_ATTRIBUTE_LABEL_KEYS} from './csvAttributeLabels'

// Read the raw file: the i18n vite plugin compiles imported message json into AST objects.
const en = JSON.parse(readFileSync(resolve(__dirname, '../../i18n/lang/en.json'), 'utf8'))

function lookup(key: string): unknown {
	return key.split('.').reduce<unknown>(
		(value, part) => (typeof value === 'object' && value !== null
			? (value as Record<string, unknown>)[part]
			: undefined),
		en,
	)
}

describe('CSV_ATTRIBUTE_LABEL_KEYS', () => {
	it.each(Object.entries(CSV_ATTRIBUTE_LABEL_KEYS))(
		'maps %s to a translation key present in en.json',
		(_attribute, key) => {
			expect(lookup(key)).toBeTypeOf('string')
		},
	)
})
