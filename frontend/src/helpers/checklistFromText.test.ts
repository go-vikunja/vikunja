import {describe, it, expect} from 'vitest'

import {findCheckboxesInText, getChecklistStatistics, getCheckboxesWithIds} from './checklistFromText'

describe('Find checklists in text', () => {
	it('should find no checkbox', () => {
		const text: string = 'Lorem Ipsum'
		const checkboxes = findCheckboxesInText(text)
		
		expect(checkboxes).toHaveLength(0)
	})
	it('should find multiple checkboxes', () => {
		const text: string = `
<ul data-type="taskList">
	<li data-checked="false" data-type="taskItem"><label><input type="checkbox"><span></span></label>
		<div><p>Task</p></div>
	</li>
	<li data-checked="false" data-type="taskItem"><label><input type="checkbox"><span></span></label>
		<div><p>Another task</p>
			<ul data-type="taskList">
				<li data-checked="false" data-type="taskItem"><label><input type="checkbox"><span></span></label>
					<div><p>subtask</p></div>
				</li>
				<li data-checked="true" data-type="taskItem"><label><input type="checkbox"
																		   checked="checked"><span></span></label>
					<div><p>done</p></div>
				</li>
			</ul>
		</div>
	</li>
</ul>`
		const checkboxes = findCheckboxesInText(text)

		expect(checkboxes).toHaveLength(4)
		expect(checkboxes[0]).toBe(32)
		expect(checkboxes[1]).toBe(163)
		expect(checkboxes[2]).toBe(321)
		expect(checkboxes[3]).toBe(464)
	})
	it('should find one unchecked checkbox', () => {
		const text: string = `
<ul data-type="taskList">
	<li data-checked="false" data-type="taskItem"><label><input type="checkbox"><span></span></label>
		<div><p>Task</p></div>
	</li>
</ul>`
		const checkboxes = findCheckboxesInText(text)

		expect(checkboxes).toHaveLength(1)
		expect(checkboxes[0]).toBe(32)
	})
	it('should find one checked checkbox', () => {
		const text: string = `
<ul data-type="taskList">
	<li data-checked="true" data-type="taskItem"><label><input type="checkbox"><span></span></label>
		<div><p>Task</p></div>
	</li>
</ul>`
		const checkboxes = findCheckboxesInText(text)

		expect(checkboxes).toHaveLength(1)
		expect(checkboxes[0]).toBe(32)
	})
})

describe('Get Checklist Statistics in a Text', () => {
	it('should find no checkbox', () => {
		const text: string = 'Lorem Ipsum'
		const stats = getChecklistStatistics(text)

		expect(stats.total).toBe(0)
	})
	it('should find one checkbox', () => {
		const text: string = `
<ul data-type="taskList">
	<li data-checked="false" data-type="taskItem"><label><input type="checkbox"><span></span></label>
		<div><p>Task</p></div>
	</li>
</ul>`
		const stats = getChecklistStatistics(text)

		expect(stats.total).toBe(1)
		expect(stats.checked).toBe(0)
	})
	it('should find one checked checkbox', () => {
		const text: string = `
<ul data-type="taskList">
	<li data-checked="true" data-type="taskItem"><label><input type="checkbox"><span></span></label>
		<div><p>Task</p></div>
	</li>
</ul>`
		const stats = getChecklistStatistics(text)

		expect(stats.total).toBe(1)
		expect(stats.checked).toBe(1)
	})
	it('should find multiple mixed and matched', () => {
		const text: string = `
<ul data-type="taskList">
	<li data-checked="false" data-type="taskItem"><label><input type="checkbox"><span></span></label>
		<div><p>Task</p></div>
	</li>
	<li data-checked="false" data-type="taskItem"><label><input type="checkbox"><span></span></label>
		<div><p>Another task</p>
			<ul data-type="taskList">
				<li data-checked="false" data-type="taskItem"><label><input type="checkbox"><span></span></label>
					<div><p>subtask</p></div>
				</li>
				<li data-checked="false" data-type="taskItem"><label><input type="checkbox"><span></span></label>
					<div><p>subtask 2</p></div>
				</li>
				<li data-checked="true" data-type="taskItem"><label><input type="checkbox"
																		   checked="checked"><span></span></label>
					<div><p>done</p></div>
				</li>
				<li data-checked="true" data-type="taskItem"><label><input type="checkbox"
																		   checked="checked"><span></span></label>
					<div><p>also done</p></div>
				</li>
			</ul>
		</div>
	</li>
</ul>`

		const stats = getChecklistStatistics(text)

		expect(stats.total).toBe(6)
		expect(stats.checked).toBe(2)
	})
})

describe('Get Checkboxes With IDs', () => {
	it('should extract checkbox info with task IDs', () => {
		const text = `
<ul data-type="taskList">
	<li data-checked="false" data-task-id="abc123"><p>Task 1</p></li>
	<li data-checked="true" data-task-id="def456"><p>Task 2</p></li>
</ul>`
		const checkboxes = getCheckboxesWithIds(text)

		expect(checkboxes).toHaveLength(2)
		expect(checkboxes[0].checked).toBe(false)
		expect(checkboxes[0].taskId).toBe('abc123')
		expect(checkboxes[1].checked).toBe(true)
		expect(checkboxes[1].taskId).toBe('def456')
	})

	it('should handle checkboxes without task IDs', () => {
		const text = `
<ul data-type="taskList">
	<li data-checked="false"><p>Legacy task</p></li>
</ul>`
		const checkboxes = getCheckboxesWithIds(text)

		expect(checkboxes).toHaveLength(1)
		expect(checkboxes[0].taskId).toBe(null)
	})
})
