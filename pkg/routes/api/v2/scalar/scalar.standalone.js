// Placeholder — the actual Scalar standalone bundle is vendored by
// `mage generate:scalarBundle` (see magefile.go) and lands in a follow-up
// PR to keep this one reviewable. Until then, /api/v2/docs renders the
// HTML shell but the Scalar reference itself will not initialise.
(function () {
	var el = document.getElementById('api-reference');
	if (el && el.parentNode) {
		var note = document.createElement('p');
		note.style.cssText = 'font-family:system-ui;padding:2rem;color:#555';
		note.textContent = 'Scalar bundle not vendored yet — run `mage generate:scalarBundle` to fetch it.';
		el.parentNode.insertBefore(note, el);
	}
})();
