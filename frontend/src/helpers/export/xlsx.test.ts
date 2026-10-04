import {describe, expect, it} from 'vitest'

import {crc32, createZip} from './zip'
import {buildXlsx, columnName, escapeXml, excelSerial, sheetNames} from './xlsx'

const decoder = new TextDecoder()

// Reads the stored entries back out of a ZIP made by createZip.
function readZip(bytes: Uint8Array): Map<string, string> {
	const view = new DataView(bytes.buffer, bytes.byteOffset, bytes.byteLength)
	const entries = new Map<string, string>()

	const eocd = bytes.length - 22
	expect(view.getUint32(eocd, true)).toBe(0x06054b50)
	const count = view.getUint16(eocd + 10, true)
	let pos = view.getUint32(eocd + 16, true)

	for (let i = 0; i < count; i++) {
		expect(view.getUint32(pos, true)).toBe(0x02014b50)
		const crc = view.getUint32(pos + 16, true)
		const size = view.getUint32(pos + 24, true)
		const nameLength = view.getUint16(pos + 28, true)
		const localOffset = view.getUint32(pos + 42, true)
		const name = decoder.decode(bytes.subarray(pos + 46, pos + 46 + nameLength))

		expect(view.getUint32(localOffset, true)).toBe(0x04034b50)
		const localName = view.getUint16(localOffset + 26, true)
		const dataStart = localOffset + 30 + localName
		const data = bytes.subarray(dataStart, dataStart + size)
		expect(crc32(data)).toBe(crc)

		entries.set(name, decoder.decode(data))
		pos += 46 + nameLength
	}
	return entries
}

describe('crc32', () => {
	it('matches the well-known check values', () => {
		expect(crc32(new TextEncoder().encode('123456789'))).toBe(0xcbf43926)
		expect(crc32(new Uint8Array())).toBe(0)
	})
})

describe('createZip', () => {
	it('round-trips entries, names and their checksums', () => {
		const enc = new TextEncoder()
		const zip = createZip([
			{name: 'a.txt', data: enc.encode('hello')},
			{name: 'dir/b.txt', data: enc.encode('世界')},
			{name: 'empty', data: new Uint8Array()},
		])

		const files = readZip(zip)
		expect([...files.keys()]).toEqual(['a.txt', 'dir/b.txt', 'empty'])
		expect(files.get('a.txt')).toBe('hello')
		expect(files.get('dir/b.txt')).toBe('世界')
		expect(files.get('empty')).toBe('')
	})

	it('starts with the local file signature, as every ZIP reader expects', () => {
		const zip = createZip([{name: 'x', data: new Uint8Array([1])}])
		expect([...zip.subarray(0, 4)]).toEqual([0x50, 0x4b, 0x03, 0x04])
	})
})

describe('columnName', () => {
	it('counts like a spreadsheet', () => {
		expect(columnName(0)).toBe('A')
		expect(columnName(25)).toBe('Z')
		expect(columnName(26)).toBe('AA')
		expect(columnName(51)).toBe('AZ')
		expect(columnName(52)).toBe('BA')
		expect(columnName(701)).toBe('ZZ')
		expect(columnName(702)).toBe('AAA')
	})
})

describe('excelSerial', () => {
	it('uses the 1900 date system', () => {
		expect(excelSerial(new Date(2026, 0, 1))).toBe(46023)
		expect(excelSerial(new Date(1970, 0, 1))).toBe(25569)
		expect(excelSerial(new Date(2026, 0, 1, 12, 0))).toBe(46023.5)
	})
})

describe('escapeXml', () => {
	it('escapes markup and removes characters XML forbids', () => {
		expect(escapeXml('a < b & "c" > d')).toBe('a &lt; b &amp; &quot;c&quot; &gt; d')
		expect(escapeXml('x\u0000y\u0008z\u000Bw')).toBe('xyzw')
		expect(escapeXml('tab\there\nnewline')).toBe('tab\there\nnewline')
	})
})

describe('sheetNames', () => {
	it('cleans, shortens and de-duplicates', () => {
		expect(sheetNames(['Tasks', 'tasks', 'A/B:C*D?[E]'])).toEqual(['Tasks', 'tasks (2)', 'A B C D  E'])
		expect(sheetNames(['x'.repeat(40)])[0]).toHaveLength(31)
		expect(sheetNames([''])[0]).toBe('Sheet')
		expect(sheetNames(["'quoted'"])[0]).toBe('quoted')
	})

	it('keeps de-duplicated names within 31 characters', () => {
		const long = 'x'.repeat(40)
		const [a, b] = sheetNames([long, long])
		expect(a).toHaveLength(31)
		expect(b.length).toBeLessThanOrEqual(31)
		expect(b).not.toBe(a)
	})
})

describe('buildXlsx', () => {
	const sheet = {
		name: 'Tasks',
		header: ['Title', 'Done', 'Due', '% done', 'Count'],
		widths: [30, 8],
		rows: [
			['Write <report> & send', true, new Date(2026, 0, 1), {value: 0.5, style: 'percent' as const}, 3],
			['=SUM(A1)', false, null, 0, undefined],
		],
	}

	it('contains every part of a spreadsheet package', () => {
		const files = readZip(buildXlsx([sheet]))
		expect([...files.keys()].sort()).toEqual([
			'[Content_Types].xml',
			'_rels/.rels',
			'xl/_rels/workbook.xml.rels',
			'xl/styles.xml',
			'xl/workbook.xml',
			'xl/worksheets/sheet1.xml',
		])
	})

	it('writes text escaped, formulas as plain text, dates as serial numbers', () => {
		const xml = readZip(buildXlsx([sheet])).get('xl/worksheets/sheet1.xml') as string

		expect(xml).toContain('Write &lt;report&gt; &amp; send')
		expect(xml).toContain('t="inlineStr"')
		expect(xml).not.toContain('<f>')
		expect(xml).toContain('>=SUM(A1)</t>')
		expect(xml).toContain('<c r="C2" s="2"><v>46023</v></c>')
		expect(xml).toContain('<c r="B2" t="b"><v>1</v></c>')
		expect(xml).toContain('<c r="B3" t="b"><v>0</v></c>')
		expect(xml).toContain('<c r="D2" s="4"><v>0.5</v></c>')
	})

	it('skips empty cells and keeps the others at their column', () => {
		const xml = readZip(buildXlsx([sheet])).get('xl/worksheets/sheet1.xml') as string

		expect(xml).not.toContain('r="C3"')
		expect(xml).not.toContain('r="E3"')
		expect(xml).toContain('<c r="D3"><v>0</v></c>')
	})

	it('styles and freezes the header, and sets the widths', () => {
		const xml = readZip(buildXlsx([sheet])).get('xl/worksheets/sheet1.xml') as string

		expect(xml).toContain('<c r="A1" t="inlineStr" s="1">')
		expect(xml).toContain('<pane ySplit="1" topLeftCell="A2" activePane="bottomLeft" state="frozen"/>')
		expect(xml).toContain('<col min="1" max="1" width="30" customWidth="1"/>')
		expect(xml).toContain('<col min="3" max="3" width="14" customWidth="1"/>')
	})

	it('can freeze leading columns too', () => {
		const xml = readZip(buildXlsx([{...sheet, freezeColumns: 2}])).get('xl/worksheets/sheet1.xml') as string
		expect(xml).toContain('<pane xSplit="2" ySplit="1" topLeftCell="C2" activePane="bottomRight" state="frozen"/>')
	})

	it('paints styled empty cells (the bars of the timeline) and ignores bad numbers and dates', () => {
		const xml = readZip(buildXlsx([{
			name: 'T',
			header: ['A', 'B', 'C', 'D'],
			rows: [[{value: null, style: 'barCritical'}, Number.NaN, new Date('nope'), Infinity]],
		}])).get('xl/worksheets/sheet1.xml') as string

		expect(xml).toContain('<c r="A2" s="8"/>')
		expect(xml).not.toContain('r="B2"')
		expect(xml).not.toContain('r="C2"')
		expect(xml).not.toContain('r="D2"')
	})

	it('lists several sheets in the workbook with unique names', () => {
		const files = readZip(buildXlsx([sheet, {...sheet, name: 'tasks'}]))
		const workbook = files.get('xl/workbook.xml') as string

		expect(workbook).toContain('<sheet name="Tasks" sheetId="1" r:id="rId1"/>')
		expect(workbook).toContain('<sheet name="tasks (2)" sheetId="2" r:id="rId2"/>')
		expect(files.has('xl/worksheets/sheet2.xml')).toBe(true)
		expect(files.get('xl/_rels/workbook.xml.rels')).toContain('Id="rId3" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles"')
	})

	it('refuses an empty workbook', () => {
		expect(() => buildXlsx([])).toThrow()
	})

	it('declares the number formats and the fills it uses', () => {
		const styles = readZip(buildXlsx([sheet])).get('xl/styles.xml') as string
		expect(styles).toContain('<cellXfs count="13">')
		expect(styles).toContain('numFmtId="164"')
		expect(styles).toContain('<fills count="11">')
	})

	it('writes text in any script', () => {
		const xml = readZip(buildXlsx([{name: 'T', header: ['名稱'], rows: [['資訊科技部 · José']]}])).get('xl/worksheets/sheet1.xml') as string
		expect(xml).toContain('名稱')
		expect(xml).toContain('資訊科技部 · José')
	})
})
