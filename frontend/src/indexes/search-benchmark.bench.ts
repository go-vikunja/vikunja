import {bench, describe} from 'vitest'
import {Document} from 'flexsearch'

interface Label {
	id: number
	title: string
	description: string
}

// Generate realistic test data
function generateLabels(count: number): Label[] {
	const words = [
		'Bug', 'Feature', 'Enhancement', 'Critical', 'Low Priority',
		'Documentation', 'Frontend', 'Backend', 'Refactor', 'Testing',
		'UI/UX', 'Performance', 'Security', 'Deployment', 'Database',
		'Blocked', 'In Progress', 'Review', 'Done', 'Backlog',
	]
	const emojis = ['🐛', '✨', '🔥', '📝', '🎨', '♻️', '🚀', '🔒', '⚡', '🐱']

	return Array.from({length: count}, (_, i) => ({
		id: i + 1,
		title: i < count * 0.1
			? emojis[i % emojis.length] // 10% emoji-only labels
			: `${words[i % words.length]} ${Math.floor(i / words.length) || ''}`.trim(),
		description: `Description for label ${i + 1}`,
	}))
}

// ---------- FlexSearch setup ----------

function setupFlexSearch(labels: Label[]) {
	const index = new Document<Label>({
		tokenize: 'full',
		document: {
			id: 'id',
			index: ['title', 'description'],
		},
	})
	for (const label of labels) {
		index.add(label.id, label)
	}
	return index
}

function flexSearchFind(index: Document<Label>, query: string): number[] {
	return index.search(query)
		?.flatMap(r => r.result)
		.filter((value, i, self) => self.indexOf(value) === i) as number[] || []
}

// ---------- String.includes() setup ----------

function includesFind(labels: Label[], query: string): number[] {
	const q = query.toLowerCase()
	return labels
		.filter(l => l.title.toLowerCase().includes(q) || l.description.toLowerCase().includes(q))
		.map(l => l.id)
}

// ---------- Benchmarks ----------

for (const size of [50, 200, 500]) {
	describe(`${size} labels`, () => {
		const labels = generateLabels(size)
		const flexIndex = setupFlexSearch(labels)

		// --- Text queries ---

		bench('FlexSearch - text query "Bug"', () => {
			flexSearchFind(flexIndex, 'Bug')
		})

		bench('String.includes - text query "Bug"', () => {
			includesFind(labels, 'Bug')
		})

		bench('FlexSearch - partial query "Enh"', () => {
			flexSearchFind(flexIndex, 'Enh')
		})

		bench('String.includes - partial query "Enh"', () => {
			includesFind(labels, 'Enh')
		})

		// --- Emoji queries ---

		bench('FlexSearch - emoji query "🐛"', () => {
			flexSearchFind(flexIndex, '🐛')
		})

		bench('String.includes - emoji query "🐛"', () => {
			includesFind(labels, '🐛')
		})

		// --- No-match queries ---

		bench('FlexSearch - no match "zzzzz"', () => {
			flexSearchFind(flexIndex, 'zzzzz')
		})

		bench('String.includes - no match "zzzzz"', () => {
			includesFind(labels, 'zzzzz')
		})
	})
}
