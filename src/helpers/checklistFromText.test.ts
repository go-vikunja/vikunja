import {findCheckboxesInText, getChecklistStatistics} from './checklistFromText'

describe('Find checklists in text', () => {
	it('should find no checkbox', () => {
		const text: string = 'Lorem Ipsum'
		const checkboxes = findCheckboxesInText(text)
		
		expect(checkboxes).toHaveLength(0)
	})
	it('should find multiple checkboxes', () => {
		const text: string = `* [ ] Lorem Ipsum
* [ ] Dolor sit amet

Here's some text in between

* [x] Dolor sit amet
- [ ] Dolor sit amet`
		const checkboxes = findCheckboxesInText(text)

		expect(checkboxes).toHaveLength(4)
		expect(checkboxes[0]).toBe(0)
		expect(checkboxes[1]).toBe(18)
		expect(checkboxes[2]).toBe(69)
		expect(checkboxes[3]).toBe(90)
	})
	it('should find one checkbox with *', () => {
		const text: string = '* [ ] Lorem Ipsum'
		const checkboxes = findCheckboxesInText(text)

		expect(checkboxes).toHaveLength(1)
		expect(checkboxes[0]).toBe(0)
	})
	it('should find one checkbox with -', () => {
		const text: string = '- [ ] Lorem Ipsum'
		const checkboxes = findCheckboxesInText(text)

		expect(checkboxes).toHaveLength(1)
		expect(checkboxes[0]).toBe(0)
	})
	it('should find one checked checkbox with *', () => {
		const text: string = '* [x] Lorem Ipsum'
		const checkboxes = findCheckboxesInText(text)

		expect(checkboxes).toHaveLength(1)
		expect(checkboxes[0]).toBe(0)
	})
	it('should find one checked checkbox with -', () => {
		const text: string = '- [x] Lorem Ipsum'
		const checkboxes = findCheckboxesInText(text)

		expect(checkboxes).toHaveLength(1)
		expect(checkboxes[0]).toBe(0)
	})
})

describe('Get Checklist Statistics in a Text', () => {
	it('should find no checkbox', () => {
		const text: string = 'Lorem Ipsum'
		const stats = getChecklistStatistics(text)

		expect(stats.total).toBe(0)
	})
	it('should find one checkbox', () => {
		const text: string = '* [ ] Lorem Ipsum'
		const stats = getChecklistStatistics(text)

		expect(stats.total).toBe(1)
		expect(stats.checked).toBe(0)
	})
	it('should find one checked checkbox', () => {
		const text: string = '* [x] Lorem Ipsum'
		const stats = getChecklistStatistics(text)

		expect(stats.total).toBe(1)
		expect(stats.checked).toBe(1)
	})
	it('should find multiple mixed and matched', () => {
		const text: string = `* [ ] Lorem Ipsum
* [ ] Dolor sit amet
* [x] Dolor sit amet
- [x] Dolor sit amet

Here's some text in between

* [x] Dolor sit amet
- [ ] Dolor sit amet`
		const stats = getChecklistStatistics(text)

		expect(stats.total).toBe(6)
		expect(stats.checked).toBe(3)
	})
})
