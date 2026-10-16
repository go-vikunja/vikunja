export const NUMBER_WORDS: Record<string, number> = {
	one: 1,
	two: 2,
	three: 3,
	four: 4,
	five: 5,
	six: 6,
	seven: 7,
	eight: 8,
	nine: 9,
	ten: 10,
	eleven: 11,
	twelve: 12,
}

export const numberWordsRegexGroup = `(${Object.keys(NUMBER_WORDS).join('|')})`

export const parseAmount = (amount: string): number => NUMBER_WORDS[amount.toLowerCase()] ?? parseInt(amount)
