/**
 * This function replaces all text, no matter the case.
 *
 * See https://stackoverflow.com/a/7313467/10924593
 *
 * @parma str
 * @param search
 * @param replace
 * @returns {*}
 */
export const replaceAll = (str, search, replace) => {
	const esc = search.replace(/[-/\\^$*+?.()|[\]{}]/g, '\\$&');
	const reg = new RegExp(esc, 'ig');
	return str.replace(reg, replace);
}