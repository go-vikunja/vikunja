export function setTitle(title) {
	document.title = (typeof title === 'undefined' || title === '')
		? 'Vikunja'
		: `${title} | Vikunja`
}