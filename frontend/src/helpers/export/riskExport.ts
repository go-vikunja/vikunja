import type {Risk} from '@/client/queries/risks'
import {parseDateOrNull} from '@/helpers/parseDateOrNull'
import type {RiskRating} from '@/helpers/riskRating'
import type {XlsxCell, XlsxSheet, XlsxStyle} from './xlsx'

export type RiskColumnKey =
	| 'id' | 'project' | 'title' | 'category' | 'status' | 'probability' | 'impact' | 'score' | 'rating' | 'owner'
	| 'identified' | 'due' | 'closed' | 'resolution' | 'mitigation' | 'contingency' | 'description'
	| 'createdBy' | 'created' | 'updated'

export interface RiskColumn {
	key: RiskColumnKey
	width: number
}

// The order of the columns of an export and of the print view.
export const ALL_RISK_COLUMNS: RiskColumn[] = [
	{key: 'id', width: 8},
	{key: 'project', width: 22},
	{key: 'title', width: 40},
	{key: 'category', width: 16},
	{key: 'status', width: 12},
	{key: 'probability', width: 11},
	{key: 'impact', width: 8},
	{key: 'score', width: 8},
	{key: 'rating', width: 10},
	{key: 'owner', width: 20},
	{key: 'identified', width: 12},
	{key: 'due', width: 12},
	{key: 'closed', width: 12},
	{key: 'resolution', width: 30},
	{key: 'mitigation', width: 36},
	{key: 'contingency', width: 36},
	{key: 'description', width: 40},
	{key: 'createdBy', width: 20},
	{key: 'created', width: 16},
	{key: 'updated', width: 16},
]

// The columns the print view leaves out: long texts do not fit a page width.
export const PRINT_RISK_COLUMNS: RiskColumnKey[] = [
	'id', 'project', 'title', 'category', 'status', 'probability', 'impact', 'score', 'rating', 'owner', 'due', 'closed',
]

export interface RiskExportLookups {
	projectTitle: (projectId: number) => string
	statusLabel: (status: Risk['status']) => string
	ratingLabel: (rating: RiskRating) => string
}

export interface RiskExportLabels {
	columns: Record<RiskColumnKey, string>
	sheet: string
}

// How a value is written on paper: readable dates in the browser's locale.
export interface RiskPrintFormatters {
	date: (d: Date) => string
	dateTime: (d: Date) => string
}

// The project column only makes sense where risks of more than one project are listed together.
export function riskColumnsFor(multipleProjects: boolean, only?: RiskColumnKey[]): RiskColumn[] {
	return ALL_RISK_COLUMNS.filter(column => {
		if (column.key === 'project' && !multipleProjects) {
			return false
		}
		return !only || only.includes(column.key)
	})
}

function userName(user: {name?: string, username?: string} | null | undefined): string {
	return user ? (user.name || user.username || '') : ''
}

const RATING_STYLE: Record<RiskRating, XlsxStyle> = {
	low: 'ratingLow',
	medium: 'ratingMedium',
	high: 'ratingHigh',
	critical: 'ratingCritical',
}

export function riskCell(risk: Risk, key: RiskColumnKey, lookups: RiskExportLookups): XlsxCell {
	switch (key) {
		case 'id': return risk.id
		case 'project': return lookups.projectTitle(risk.project_id)
		case 'title': return risk.title
		case 'category': return risk.category
		case 'status': return lookups.statusLabel(risk.status)
		case 'probability': return risk.probability
		case 'impact': return risk.impact
		case 'score': return risk.score
		case 'rating': return {value: lookups.ratingLabel(risk.rating), style: RATING_STYLE[risk.rating]}
		case 'owner': return userName(risk.owner)
		case 'identified': return {value: parseDateOrNull(risk.identified_date), style: 'date'}
		case 'due': return {value: parseDateOrNull(risk.due_date), style: 'date'}
		case 'closed': return {value: parseDateOrNull(risk.closed_at), style: 'date'}
		case 'resolution': return risk.resolution
		case 'mitigation': return risk.mitigation
		case 'contingency': return risk.contingency
		case 'description': return risk.description
		case 'createdBy': return userName(risk.created_by)
		case 'created': return {value: parseDateOrNull(risk.created), style: 'datetime'}
		case 'updated': return {value: parseDateOrNull(risk.updated), style: 'datetime'}
	}
}

export function risksSheet(
	risks: Risk[],
	multipleProjects: boolean,
	lookups: RiskExportLookups,
	labels: RiskExportLabels,
): XlsxSheet {
	const columns = riskColumnsFor(multipleProjects)
	return {
		name: labels.sheet,
		header: columns.map(c => labels.columns[c.key]),
		widths: columns.map(c => c.width),
		rows: risks.map(risk => columns.map(c => riskCell(risk, c.key, lookups))),
		// The id and the title stay in view while scrolling sideways.
		freezeColumns: columns.findIndex(c => c.key === 'title') + 1,
	}
}

// ---- Print ----------------------------------------------------------------------------------------

export interface PrintRiskTable {
	headers: string[]
	rows: string[][]
	// The rating of every row, for its colour on paper.
	ratings: RiskRating[]
	// Index of the rating column in a row, -1 if there is none.
	ratingColumn: number
	closedRows: boolean[]
}

export function riskPrintValue(
	risk: Risk,
	key: RiskColumnKey,
	lookups: RiskExportLookups,
	fmt: RiskPrintFormatters,
): string {
	const day = (v: string | null | undefined) => {
		const d = parseDateOrNull(v)
		return d ? fmt.date(d) : ''
	}
	const dayTime = (v: string | null | undefined) => {
		const d = parseDateOrNull(v)
		return d ? fmt.dateTime(d) : ''
	}
	switch (key) {
		case 'id': return String(risk.id)
		case 'project': return lookups.projectTitle(risk.project_id)
		case 'title': return risk.title
		case 'category': return risk.category
		case 'status': return lookups.statusLabel(risk.status)
		case 'probability': return String(risk.probability)
		case 'impact': return String(risk.impact)
		case 'score': return String(risk.score)
		case 'rating': return lookups.ratingLabel(risk.rating)
		case 'owner': return userName(risk.owner)
		case 'identified': return day(risk.identified_date)
		case 'due': return day(risk.due_date)
		case 'closed': return day(risk.closed_at)
		case 'resolution': return risk.resolution
		case 'mitigation': return risk.mitigation
		case 'contingency': return risk.contingency
		case 'description': return risk.description
		case 'createdBy': return userName(risk.created_by)
		case 'created': return dayTime(risk.created)
		case 'updated': return dayTime(risk.updated)
	}
}

export function buildRiskPrintTable(
	risks: Risk[],
	multipleProjects: boolean,
	lookups: RiskExportLookups,
	fmt: RiskPrintFormatters,
	labels: RiskExportLabels,
): PrintRiskTable {
	const keys = riskColumnsFor(multipleProjects, PRINT_RISK_COLUMNS).map(c => c.key)
	return {
		headers: keys.map(key => labels.columns[key]),
		rows: risks.map(risk => keys.map(key => riskPrintValue(risk, key, lookups, fmt))),
		ratings: risks.map(risk => risk.rating),
		ratingColumn: keys.indexOf('rating'),
		closedRows: risks.map(risk => risk.status === 'closed'),
	}
}

// <scope>-<yyyy-mm-dd>.<ext>, safe on every file system. Same rules as the task exports.
export {exportFileName as riskExportFileName} from './taskExport'
