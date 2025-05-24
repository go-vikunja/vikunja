export const downloadBlob = (url: string, filename: string) => {
	const link = document.createElement('a')
	link.href = url
	link.setAttribute('download', filename)
	link.click()
	window.URL.revokeObjectURL(url)
}
