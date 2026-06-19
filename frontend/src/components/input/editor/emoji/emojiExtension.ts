import {Extension} from '@tiptap/core'
import Suggestion from '@tiptap/suggestion'

import emojiSuggestionSetup from './emojiSuggestion'

export const EmojiExtension = Extension.create({
	name: 'emojiAutocomplete',

	addOptions() {
		return {
			suggestion: emojiSuggestionSetup(),
		}
	},

	addProseMirrorPlugins() {
		return [
			Suggestion({
				editor: this.editor,
				...this.options.suggestion,
			}),
		]
	},
})
