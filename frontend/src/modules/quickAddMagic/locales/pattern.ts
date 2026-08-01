/**
 * Escapes a literal form for use in a regex. Forms containing a character
 * class (e.g. Ukrainian apostrophes in п['’]ятниця) are already valid regex
 * fragments and are used as-is.
 */
export const toPattern = (form: string): string =>
	form.includes('[') ? form : form.replace(/[-/\\^$*+?.()|[\]{}]/g, '\\$&')

/** Finds the dictionary entry whose form produced the given regex match. */
export const findByForm = <T extends { forms: string[] }>(defs: T[], matched: string): T | undefined =>
	defs.find(d => d.forms.some(f => new RegExp(`^${toPattern(f)}$`, 'i').test(matched)))
