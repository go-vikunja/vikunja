// Markup markdown conversion would throw away. Code editors put syntax-highlighted
// span/div soup on the clipboard, so the presence of text/html alone says nothing
// about whether the copied text was meant as markdown source.
const FORMATTING_SELECTOR = 'a,strong,b,em,i,u,s,strike,del,ins,mark,small,code,pre,kbd,sub,sup,h1,h2,h3,h4,h5,h6,ul,ol,li,blockquote,table,img,hr,figure'

export function htmlHasFormatting(html: string): boolean {
	if (!html) {
		return false
	}

	const doc = new DOMParser().parseFromString(html, 'text/html')

	return doc.body.querySelector(FORMATTING_SELECTOR) !== null
}
