import { Schema } from '@tiptap/pm/model'

// Define the schema for our filter editor
export const schema = new Schema({
	nodes: {
		doc: {
			content: 'paragraph+'
		},
		paragraph: {
			content: 'text*',
			toDOM() { return ['p', 0] }
		},
		text: {}
	},
	marks: {
		// Marks for different parts of the filter expression
		field: {
			toDOM() { return ['span', { class: 'field' }, 0] },
			parseDOM: [{ tag: 'span.field' }]
		},
		operator: {
			toDOM() { return ['span', { class: 'operator' }, 0] },
			parseDOM: [{ tag: 'span.operator' }]
		},
		value: {
			toDOM() { return ['span', { class: 'value' }, 0] },
			parseDOM: [{ tag: 'span.value' }]
		},
		logical: {
			toDOM() { return ['span', { class: 'logical' }, 0] },
			parseDOM: [{ tag: 'span.logical' }]
		},
		grouping: {
			toDOM() { return ['span', { class: 'grouping' }, 0] },
			parseDOM: [{ tag: 'span.grouping' }]
		}
	}
});
