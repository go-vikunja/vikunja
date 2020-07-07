
export const setTitle = title => {
	if (typeof title === 'undefined' || title === '') {
		document.title = 'Vikunja'
		return
	}

	document.title = `${title} | Vikunja`
}