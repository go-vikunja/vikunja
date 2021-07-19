export const getNamespaceTitle = (n, $t) => {
	if (n.id === -1) {
		return $t('namespace.pseudo.sharedLists.title')
	}
	if (n.id === -2) {
		return $t('namespace.pseudo.favorites.title')
	}
	if (n.id === -3) {
		return $t('namespace.pseudo.savedFilters.title')
	}
	return n.title
}
