import {createZip, type ZipEntry} from './zip'

// A small, dependency-free .xlsx writer: text, numbers, booleans and dates, a styled and frozen
// header row, column widths and a few fills (used by the Gantt timeline sheet).

export type XlsxStyle =
	| 'header'
	| 'date'
	| 'datetime'
	| 'percent'
	| 'barTask'
	| 'barSummary'
	| 'barMilestone'
	| 'barCritical'

export type XlsxValue = string | number | boolean | Date | null | undefined

export interface XlsxStyledCell {
	value: XlsxValue
	style?: XlsxStyle
}

export type XlsxCell = XlsxValue | XlsxStyledCell

export interface XlsxSheet {
	name: string
	// The header row, written first, styled and frozen.
	header: string[]
	// Width in characters per column, a default is used where missing.
	widths?: number[]
	rows: XlsxCell[][]
	// Also freeze this many columns on the left.
	freezeColumns?: number
}

const STYLE_INDEX: Record<XlsxStyle, number> = {
	header: 1,
	date: 2,
	datetime: 3,
	percent: 4,
	barTask: 5,
	barSummary: 6,
	barMilestone: 7,
	barCritical: 8,
}

export const MAX_SHEET_COLUMNS = 16384
const DEFAULT_WIDTH = 14
const encoder = new TextEncoder()

// Characters XML 1.0 does not allow at all. They are dropped instead of making the file unreadable.
// eslint-disable-next-line no-control-regex
const INVALID_XML_CHARS = /[\u0000-\u0008\u000B\u000C\u000E-\u001F\uFFFE\uFFFF]/g

export function escapeXml(text: string): string {
	return text
		.replace(INVALID_XML_CHARS, '')
		.replace(/&/g, '&amp;')
		.replace(/</g, '&lt;')
		.replace(/>/g, '&gt;')
		.replace(/"/g, '&quot;')
}

// 0 -> A, 25 -> Z, 26 -> AA
export function columnName(index: number): string {
	let n = index
	let name = ''
	do {
		name = String.fromCharCode(65 + (n % 26)) + name
		n = Math.floor(n / 26) - 1
	} while (n >= 0)
	return name
}

// Excel stores a date as the days since 1899-12-30. The local calendar time is used, which is what
// the user sees on screen.
export function excelSerial(date: Date): number {
	const utc = Date.UTC(date.getFullYear(), date.getMonth(), date.getDate(), date.getHours(), date.getMinutes(), date.getSeconds())
	return utc / 86_400_000 + 25569
}

// Sheet names: at most 31 characters, none of []:*?/\, not empty, unique ignoring case.
export function sheetNames(requested: string[]): string[] {
	const used = new Set<string>()
	return requested.map(raw => {
		let base = raw.replace(/[[\]:*?/\\]/g, ' ').replace(/^'+|'+$/g, '').trim().slice(0, 31).trim()
		if (base === '') {
			base = 'Sheet'
		}
		let name = base
		for (let i = 2; used.has(name.toLowerCase()); i++) {
			const suffix = ` (${i})`
			name = base.slice(0, 31 - suffix.length).trimEnd() + suffix
		}
		used.add(name.toLowerCase())
		return name
	})
}

function cellXml(ref: string, cell: XlsxCell, forcedStyle?: XlsxStyle): string {
	let value: XlsxValue
	let style: XlsxStyle | undefined = forcedStyle
	if (cell !== null && typeof cell === 'object' && !(cell instanceof Date)) {
		value = cell.value
		style = cell.style ?? style
	} else {
		value = cell
	}

	if (value === null || value === undefined || value === '') {
		// A styled empty cell still paints its fill.
		return style ? `<c r="${ref}" s="${STYLE_INDEX[style]}"/>` : ''
	}

	if (value instanceof Date) {
		if (Number.isNaN(value.getTime())) {
			return ''
		}
		const s = STYLE_INDEX[style === 'datetime' ? 'datetime' : style ?? 'date']
		return `<c r="${ref}" s="${s}"><v>${excelSerial(value)}</v></c>`
	}
	if (typeof value === 'number') {
		if (!Number.isFinite(value)) {
			return ''
		}
		return `<c r="${ref}"${style ? ` s="${STYLE_INDEX[style]}"` : ''}><v>${value}</v></c>`
	}
	if (typeof value === 'boolean') {
		return `<c r="${ref}" t="b"${style ? ` s="${STYLE_INDEX[style]}"` : ''}><v>${value ? 1 : 0}</v></c>`
	}

	// Inline strings are always text: Excel never evaluates them as a formula, so a title like
	// "=SUM(A1)" stays what it is and does not need escaping against formula injection.
	return `<c r="${ref}" t="inlineStr"${style ? ` s="${STYLE_INDEX[style]}"` : ''}><is><t xml:space="preserve">${escapeXml(value)}</t></is></c>`
}

function sheetXml(sheet: XlsxSheet): string {
	const columnCount = Math.min(MAX_SHEET_COLUMNS, Math.max(sheet.header.length, ...sheet.rows.map(r => r.length), 1))

	const cols = Array.from({length: columnCount}, (_, i) => {
		const width = sheet.widths?.[i] ?? DEFAULT_WIDTH
		return `<col min="${i + 1}" max="${i + 1}" width="${width}" customWidth="1"/>`
	}).join('')

	const rows: string[] = []
	rows.push(`<row r="1">${sheet.header
		.slice(0, columnCount)
		.map((text, i) => cellXml(`${columnName(i)}1`, text, 'header'))
		.join('')}</row>`)
	sheet.rows.forEach((row, rowIndex) => {
		const r = rowIndex + 2
		const cells = row.slice(0, columnCount).map((cell, i) => cellXml(`${columnName(i)}${r}`, cell)).join('')
		rows.push(`<row r="${r}">${cells}</row>`)
	})

	const freezeColumns = Math.max(0, Math.min(sheet.freezeColumns ?? 0, columnCount - 1))
	const topLeft = `${columnName(freezeColumns)}2`
	const pane = freezeColumns > 0
		? `<pane xSplit="${freezeColumns}" ySplit="1" topLeftCell="${topLeft}" activePane="bottomRight" state="frozen"/>`
		: `<pane ySplit="1" topLeftCell="A2" activePane="bottomLeft" state="frozen"/>`

	return '<?xml version="1.0" encoding="UTF-8" standalone="yes"?>' +
		'<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">' +
		`<sheetViews><sheetView workbookViewId="0">${pane}</sheetView></sheetViews>` +
		'<sheetFormatPr defaultRowHeight="15"/>' +
		`<cols>${cols}</cols>` +
		`<sheetData>${rows.join('')}</sheetData>` +
		'</worksheet>'
}

const STYLES_XML = '<?xml version="1.0" encoding="UTF-8" standalone="yes"?>' +
	'<styleSheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">' +
	'<numFmts count="2"><numFmt numFmtId="164" formatCode="yyyy\\-mm\\-dd"/><numFmt numFmtId="165" formatCode="yyyy\\-mm\\-dd\\ hh:mm"/></numFmts>' +
	'<fonts count="2"><font><sz val="11"/><name val="Calibri"/></font><font><b/><sz val="11"/><name val="Calibri"/></font></fonts>' +
	'<fills count="7">' +
	'<fill><patternFill patternType="none"/></fill>' +
	'<fill><patternFill patternType="gray125"/></fill>' +
	'<fill><patternFill patternType="solid"><fgColor rgb="FFD9D9D9"/><bgColor indexed="64"/></patternFill></fill>' +
	'<fill><patternFill patternType="solid"><fgColor rgb="FF3B82F6"/><bgColor indexed="64"/></patternFill></fill>' +
	'<fill><patternFill patternType="solid"><fgColor rgb="FF374151"/><bgColor indexed="64"/></patternFill></fill>' +
	'<fill><patternFill patternType="solid"><fgColor rgb="FFF59E0B"/><bgColor indexed="64"/></patternFill></fill>' +
	'<fill><patternFill patternType="solid"><fgColor rgb="FFEF4444"/><bgColor indexed="64"/></patternFill></fill>' +
	'</fills>' +
	'<borders count="1"><border><left/><right/><top/><bottom/><diagonal/></border></borders>' +
	'<cellStyleXfs count="1"><xf numFmtId="0" fontId="0" fillId="0" borderId="0"/></cellStyleXfs>' +
	'<cellXfs count="9">' +
	'<xf numFmtId="0" fontId="0" fillId="0" borderId="0" xfId="0"/>' +
	'<xf numFmtId="0" fontId="1" fillId="2" borderId="0" xfId="0" applyFont="1" applyFill="1"/>' +
	'<xf numFmtId="164" fontId="0" fillId="0" borderId="0" xfId="0" applyNumberFormat="1"/>' +
	'<xf numFmtId="165" fontId="0" fillId="0" borderId="0" xfId="0" applyNumberFormat="1"/>' +
	'<xf numFmtId="9" fontId="0" fillId="0" borderId="0" xfId="0" applyNumberFormat="1"/>' +
	'<xf numFmtId="0" fontId="0" fillId="3" borderId="0" xfId="0" applyFill="1"/>' +
	'<xf numFmtId="0" fontId="0" fillId="4" borderId="0" xfId="0" applyFill="1"/>' +
	'<xf numFmtId="0" fontId="0" fillId="5" borderId="0" xfId="0" applyFill="1"/>' +
	'<xf numFmtId="0" fontId="0" fillId="6" borderId="0" xfId="0" applyFill="1"/>' +
	'</cellXfs>' +
	'<cellStyles count="1"><cellStyle name="Normal" xfId="0" builtinId="0"/></cellStyles>' +
	'</styleSheet>'

const XML_DECL = '<?xml version="1.0" encoding="UTF-8" standalone="yes"?>'

export function buildXlsx(sheets: XlsxSheet[]): Uint8Array {
	if (sheets.length === 0) {
		throw new Error('A workbook needs at least one sheet')
	}
	const names = sheetNames(sheets.map(s => s.name))
	const entries: ZipEntry[] = []
	const add = (name: string, xml: string) => entries.push({name, data: encoder.encode(xml)})

	add('[Content_Types].xml', XML_DECL +
		'<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">' +
		'<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>' +
		'<Default Extension="xml" ContentType="application/xml"/>' +
		'<Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>' +
		sheets.map((_, i) => `<Override PartName="/xl/worksheets/sheet${i + 1}.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>`).join('') +
		'<Override PartName="/xl/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.styles+xml"/>' +
		'</Types>')

	add('_rels/.rels', XML_DECL +
		'<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">' +
		'<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/>' +
		'</Relationships>')

	add('xl/workbook.xml', XML_DECL +
		'<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">' +
		`<sheets>${names.map((name, i) => `<sheet name="${escapeXml(name)}" sheetId="${i + 1}" r:id="rId${i + 1}"/>`).join('')}</sheets>` +
		'</workbook>')

	add('xl/_rels/workbook.xml.rels', XML_DECL +
		'<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">' +
		sheets.map((_, i) => `<Relationship Id="rId${i + 1}" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet${i + 1}.xml"/>`).join('') +
		`<Relationship Id="rId${sheets.length + 1}" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>` +
		'</Relationships>')

	add('xl/styles.xml', STYLES_XML)
	sheets.forEach((sheet, i) => add(`xl/worksheets/sheet${i + 1}.xml`, sheetXml(sheet)))

	return createZip(entries)
}

export const XLSX_MIME = 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet'
