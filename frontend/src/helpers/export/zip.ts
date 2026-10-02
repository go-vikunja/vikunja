// A minimal ZIP writer: entries are stored, not compressed, which is all an .xlsx needs and keeps
// this free of dependencies. Names and sizes follow the ZIP spec (APPNOTE 6.3), UTF-8 names.

export interface ZipEntry {
	name: string
	data: Uint8Array
}

let crcTable: Uint32Array | undefined

function getCrcTable(): Uint32Array {
	if (crcTable) {
		return crcTable
	}
	const table = new Uint32Array(256)
	for (let n = 0; n < 256; n++) {
		let c = n
		for (let k = 0; k < 8; k++) {
			c = c & 1 ? 0xedb88320 ^ (c >>> 1) : c >>> 1
		}
		table[n] = c >>> 0
	}
	crcTable = table
	return table
}

export function crc32(data: Uint8Array): number {
	const table = getCrcTable()
	let crc = 0xffffffff
	for (let i = 0; i < data.length; i++) {
		crc = table[(crc ^ data[i]) & 0xff] ^ (crc >>> 8)
	}
	return (crc ^ 0xffffffff) >>> 0
}

// 2026-01-01 00:00:00 in MS-DOS date and time, entries do not need a real timestamp.
const DOS_DATE = ((2026 - 1980) << 9) | (1 << 5) | 1
const DOS_TIME = 0
const UTF8_FLAG = 0x0800

const encoder = new TextEncoder()

export function createZip(entries: ZipEntry[]): Uint8Array {
	const chunks: Uint8Array[] = []
	const central: Uint8Array[] = []
	let offset = 0

	for (const entry of entries) {
		const name = encoder.encode(entry.name)
		const crc = crc32(entry.data)
		const size = entry.data.length

		const local = new DataView(new ArrayBuffer(30))
		local.setUint32(0, 0x04034b50, true)
		local.setUint16(4, 20, true) // version needed
		local.setUint16(6, UTF8_FLAG, true)
		local.setUint16(8, 0, true) // stored
		local.setUint16(10, DOS_TIME, true)
		local.setUint16(12, DOS_DATE, true)
		local.setUint32(14, crc, true)
		local.setUint32(18, size, true)
		local.setUint32(22, size, true)
		local.setUint16(26, name.length, true)
		local.setUint16(28, 0, true)

		const header = new DataView(new ArrayBuffer(46))
		header.setUint32(0, 0x02014b50, true)
		header.setUint16(4, 20, true) // version made by
		header.setUint16(6, 20, true) // version needed
		header.setUint16(8, UTF8_FLAG, true)
		header.setUint16(10, 0, true)
		header.setUint16(12, DOS_TIME, true)
		header.setUint16(14, DOS_DATE, true)
		header.setUint32(16, crc, true)
		header.setUint32(20, size, true)
		header.setUint32(24, size, true)
		header.setUint16(28, name.length, true)
		header.setUint16(30, 0, true) // extra
		header.setUint16(32, 0, true) // comment
		header.setUint16(34, 0, true) // disk
		header.setUint16(36, 0, true) // internal attributes
		header.setUint32(38, 0, true) // external attributes
		header.setUint32(42, offset, true)

		chunks.push(new Uint8Array(local.buffer), name, entry.data)
		central.push(new Uint8Array(header.buffer), name)
		offset += 30 + name.length + size
	}

	const centralSize = central.reduce((sum, c) => sum + c.length, 0)
	const end = new DataView(new ArrayBuffer(22))
	end.setUint32(0, 0x06054b50, true)
	end.setUint16(8, entries.length, true)
	end.setUint16(10, entries.length, true)
	end.setUint32(12, centralSize, true)
	end.setUint32(16, offset, true)

	const all = [...chunks, ...central, new Uint8Array(end.buffer)]
	const out = new Uint8Array(all.reduce((sum, c) => sum + c.length, 0))
	let pos = 0
	for (const chunk of all) {
		out.set(chunk, pos)
		pos += chunk.length
	}
	return out
}
