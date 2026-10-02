// Hands a file to the browser as a download.
export function downloadBytes(fileName: string, bytes: Uint8Array, mime: string) {
	// Copy into a plain ArrayBuffer: a Uint8Array may sit on a shared buffer, which Blob rejects.
	const blob = new Blob([bytes.slice().buffer], {type: mime})
	const url = URL.createObjectURL(blob)
	const link = document.createElement('a')
	link.href = url
	link.download = fileName
	link.rel = 'noopener'
	document.body.appendChild(link)
	link.click()
	link.remove()
	// Give the browser a moment to start the download before the URL goes away.
	setTimeout(() => URL.revokeObjectURL(url), 10_000)
}
