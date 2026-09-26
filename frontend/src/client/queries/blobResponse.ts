// The generated operations type a blob download as its JSON schema, so every caller has to narrow
// the parsed body itself.
export function expectBlob(data: unknown, label: string): Blob {
	if (!(data instanceof Blob)) {
		throw new Error(`${label} response was not a blob`)
	}
	return data
}
