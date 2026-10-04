import {describe, expect, it} from 'vitest'

import type {Risk} from '@/client/queries/risks'
import {
	ALL_RISK_COLUMNS,
	buildRiskPrintTable,
	PRINT_RISK_COLUMNS,
	riskCell,
	riskColumnsFor,
	risksSheet,
	type RiskColumnKey,
	type RiskExportLabels,
	type RiskExportLookups,
} from './riskExport'
import {buildXlsx} from './xlsx'

function risk(overrides: Partial<Risk> = {}): Risk {
	return {
		id: 12, project_id: 7, title: 'Supplier delay', description: 'Parts arrive late', category: 'Schedule',
		probability: 4, impact: 5, score: 20, rating: 'critical', owner_id: 3,
		owner: {id: 3, name: 'Ada Lovelace', username: 'ada'}, mitigation: 'Second supplier', contingency: 'Air freight',
		status: 'mitigating', identified_date: '2026-09-01T00:00:00Z', due_date: '2026-10-31T23:59:59Z', closed_at: null,
		closed_by: null, resolution: '', created_by: {id: 1, name: '', username: 'root'},
		created: '2026-09-01T08:00:00Z', updated: '2026-09-02T08:00:00Z',
		...overrides,
	}
}

const lookups: RiskExportLookups = {
	projectTitle: id => (id === 7 ? 'Launch' : ''),
	statusLabel: status => `status:${status}`,
	ratingLabel: rating => `rating:${rating}`,
}

const labels: RiskExportLabels = {
	sheet: 'Risks',
	columns: Object.fromEntries(ALL_RISK_COLUMNS.map(c => [c.key, `col:${c.key}`])) as Record<RiskColumnKey, string>,
}

describe('riskColumnsFor', () => {
	it('leaves the project column out for a single project', () => {
		expect(riskColumnsFor(false).map(c => c.key)).not.toContain('project')
		expect(riskColumnsFor(true).map(c => c.key)).toContain('project')
	})

	it('keeps the order of the columns and can be narrowed', () => {
		expect(riskColumnsFor(true).map(c => c.key)).toEqual(ALL_RISK_COLUMNS.map(c => c.key))
		expect(riskColumnsFor(true, ['title', 'id']).map(c => c.key)).toEqual(['id', 'title'])
	})

	it('lists the columns of the plan', () => {
		expect(ALL_RISK_COLUMNS.map(c => c.key)).toEqual([
			'id', 'project', 'title', 'category', 'status', 'probability', 'impact', 'score', 'rating', 'owner',
			'identified', 'due', 'closed', 'resolution', 'mitigation', 'contingency', 'description',
			'createdBy', 'created', 'updated',
		])
	})
})

describe('riskCell', () => {
	it('writes numbers as numbers and labels for status and rating', () => {
		expect(riskCell(risk(), 'id', lookups)).toBe(12)
		expect(riskCell(risk(), 'score', lookups)).toBe(20)
		expect(riskCell(risk(), 'status', lookups)).toBe('status:mitigating')
		expect(riskCell(risk(), 'project', lookups)).toBe('Launch')
	})

	it('colours the rating by its band', () => {
		expect(riskCell(risk({rating: 'low'}), 'rating', lookups)).toEqual({value: 'rating:low', style: 'ratingLow'})
		expect(riskCell(risk({rating: 'medium'}), 'rating', lookups)).toEqual({value: 'rating:medium', style: 'ratingMedium'})
		expect(riskCell(risk({rating: 'high'}), 'rating', lookups)).toEqual({value: 'rating:high', style: 'ratingHigh'})
		expect(riskCell(risk({rating: 'critical'}), 'rating', lookups)).toEqual({value: 'rating:critical', style: 'ratingCritical'})
	})

	it('names people by name, then by username, and leaves a missing one empty', () => {
		expect(riskCell(risk(), 'owner', lookups)).toBe('Ada Lovelace')
		expect(riskCell(risk(), 'createdBy', lookups)).toBe('root')
		expect(riskCell(risk({owner: null}), 'owner', lookups)).toBe('')
	})

	it('writes dates as dates and a missing date as an empty cell', () => {
		const due = riskCell(risk(), 'due', lookups) as {value: Date | null, style: string}
		expect(due.value).toBeInstanceOf(Date)
		expect(due.style).toBe('date')
		expect(riskCell(risk(), 'closed', lookups)).toEqual({value: null, style: 'date'})
	})
})

describe('risksSheet', () => {
	it('has a header, widths and a row per risk, in the same columns', () => {
		const sheet = risksSheet([risk(), risk({id: 13})], true, lookups, labels)
		expect(sheet.name).toBe('Risks')
		expect(sheet.header).toHaveLength(ALL_RISK_COLUMNS.length)
		expect(sheet.header[0]).toBe('col:id')
		expect(sheet.widths).toHaveLength(sheet.header.length)
		expect(sheet.rows).toHaveLength(2)
		expect(sheet.rows[0]).toHaveLength(sheet.header.length)
	})

	it('has no project column for a single project', () => {
		const sheet = risksSheet([risk()], false, lookups, labels)
		expect(sheet.header).not.toContain('col:project')
		expect(sheet.header).toHaveLength(ALL_RISK_COLUMNS.length - 1)
	})

	it('freezes the columns up to the title', () => {
		expect(risksSheet([], true, lookups, labels).freezeColumns).toBe(3)
		expect(risksSheet([], false, lookups, labels).freezeColumns).toBe(2)
	})

	it('builds a workbook, also with no risks at all', () => {
		expect(buildXlsx([risksSheet([], true, lookups, labels)]).length).toBeGreaterThan(0)
		expect(buildXlsx([risksSheet([risk()], true, lookups, labels)]).length).toBeGreaterThan(0)
	})
})

describe('buildRiskPrintTable', () => {
	const fmt = {date: (d: Date) => `d:${d.getUTCFullYear()}`, dateTime: (d: Date) => `dt:${d.getUTCFullYear()}`}

	it('prints the short columns only, with a project column when asked', () => {
		const table = buildRiskPrintTable([risk()], true, lookups, fmt, labels)
		expect(table.headers).toEqual(PRINT_RISK_COLUMNS.map(key => `col:${key}`))
		expect(table.headers).not.toContain('col:mitigation')
		const single = buildRiskPrintTable([risk()], false, lookups, fmt, labels)
		expect(single.headers).not.toContain('col:project')
	})

	it('writes values as text and remembers the rating and the closed rows', () => {
		const table = buildRiskPrintTable([risk(), risk({id: 13, status: 'closed', rating: 'low', closed_at: '2026-10-02T00:00:00Z'})], true, lookups, fmt, labels)
		expect(table.rows[0]).toEqual(['12', 'Launch', 'Supplier delay', 'Schedule', 'status:mitigating', '4', '5', '20', 'rating:critical', 'Ada Lovelace', 'd:2026', ''])
		expect(table.ratings).toEqual(['critical', 'low'])
		expect(table.closedRows).toEqual([false, true])
		expect(table.ratingColumn).toBe(table.headers.indexOf('col:rating'))
		expect(table.rows[1][table.headers.indexOf('col:closed')]).toBe('d:2026')
	})

	it('finds the rating column, which moves with the project column', () => {
		expect(buildRiskPrintTable([], false, lookups, fmt, labels).ratingColumn).toBe(7)
		expect(buildRiskPrintTable([], true, lookups, fmt, labels).ratingColumn).toBe(8)
	})
})
