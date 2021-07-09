export const getListTitle = (l, $t) => {
	if (l.id === -1) {
		return $t('list.pseudo.favorites.title');
	}
	return l.title;
}
