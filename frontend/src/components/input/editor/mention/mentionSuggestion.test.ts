import {describe, it, expect} from 'vitest'

import mentionSuggestionSetup from './mentionSuggestion'

describe('mentionSuggestion', () => {
	it('exits cleanly when the suggestion was never started', () => {
		// tiptap recreates the plugin view while a suggestion is active, so onExit can
		// run without a matching onStart
		const renderer = mentionSuggestionSetup(1).render()

		expect(() => renderer.onExit()).not.toThrow()
	})
})
