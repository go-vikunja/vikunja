import {describe, expect, it} from 'vitest'
import {parseSubtasksViaIndention} from '@/helpers/parseSubtasksViaIndention'
import {PrefixMode} from '@/modules/parseTaskText'

describe('Parse Subtasks via Relation', () => {
	it('Should not return a parent for a single task', () => {
		const tasks = parseSubtasksViaIndention('single task', PrefixMode.Default)
		
		expect(tasks).to.have.length(1)
		expect(tasks[0].parent).toBeNull()
	})
	it('Should not return a parent for multiple tasks without indention', () => {
		const tasks = parseSubtasksViaIndention(`task one
task two`, PrefixMode.Default)

		expect(tasks).to.have.length(2)
		expect(tasks[0].parent).toBeNull()
		expect(tasks[1].parent).toBeNull()
	})
	it('Should return a parent for two tasks with indention', () => {
		const tasks = parseSubtasksViaIndention(`parent task
  sub task`, PrefixMode.Default)

		expect(tasks).to.have.length(2)
		expect(tasks[0].parent).toBeNull()
		expect(tasks[0].title).to.eq('parent task')
		expect(tasks[1].parent).to.eq('parent task')
		expect(tasks[1].title).to.eq('sub task')
	})
	it('Should return a parent for multiple subtasks', () => {
		const tasks = parseSubtasksViaIndention(`parent task
  sub task one
  sub task two`, PrefixMode.Default)

		expect(tasks).to.have.length(3)
		expect(tasks[0].parent).toBeNull()
		expect(tasks[0].title).to.eq('parent task')
		expect(tasks[1].title).to.eq('sub task one')
		expect(tasks[1].parent).to.eq('parent task')
		expect(tasks[2].title).to.eq('sub task two')
		expect(tasks[2].parent).to.eq('parent task')
	})
	it('Should work with multiple indention levels', () => {
		const tasks = parseSubtasksViaIndention(`parent task
  sub task
    sub sub task`, PrefixMode.Default)

		expect(tasks).to.have.length(3)
		expect(tasks[0].parent).toBeNull()
		expect(tasks[0].title).to.eq('parent task')
		expect(tasks[1].title).to.eq('sub task')
		expect(tasks[1].parent).to.eq('parent task')
		expect(tasks[2].title).to.eq('sub sub task')
		expect(tasks[2].parent).to.eq('sub task')
	})
	it('Should work with multiple indention levels and multiple tasks', () => {
		const tasks = parseSubtasksViaIndention(`parent task
  sub task
    sub sub task one
    sub sub task two`, PrefixMode.Default)

		expect(tasks).to.have.length(4)
		expect(tasks[0].parent).toBeNull()
		expect(tasks[0].title).to.eq('parent task')
		expect(tasks[1].title).to.eq('sub task')
		expect(tasks[1].parent).to.eq('parent task')
		expect(tasks[2].title).to.eq('sub sub task one')
		expect(tasks[2].parent).to.eq('sub task')
		expect(tasks[3].title).to.eq('sub sub task two')
		expect(tasks[3].parent).to.eq('sub task')
	})
	it('Should work with multiple indention levels and multiple tasks', () => {
		const tasks = parseSubtasksViaIndention(`parent task
  sub task
    sub sub task one
      sub sub sub task
    sub sub task two`, PrefixMode.Default)

		expect(tasks).to.have.length(5)
		expect(tasks[0].parent).toBeNull()
		expect(tasks[0].title).to.eq('parent task')
		expect(tasks[1].title).to.eq('sub task')
		expect(tasks[1].parent).to.eq('parent task')
		expect(tasks[2].title).to.eq('sub sub task one')
		expect(tasks[2].parent).to.eq('sub task')
		expect(tasks[3].title).to.eq('sub sub sub task')
		expect(tasks[3].parent).to.eq('sub sub task one')
		expect(tasks[4].title).to.eq('sub sub task two')
		expect(tasks[4].parent).to.eq('sub task')
	})
	it('Should return a parent for multiple subtasks with special stuff', () => {
		const tasks = parseSubtasksViaIndention(`* parent task
  * sub task one
  sub task two`, PrefixMode.Default)

		expect(tasks).to.have.length(3)
		expect(tasks[0].parent).toBeNull()
		expect(tasks[0].title).to.eq('parent task')
		expect(tasks[1].title).to.eq('sub task one')
		expect(tasks[1].parent).to.eq('parent task')
		expect(tasks[2].title).to.eq('sub task two')
		expect(tasks[2].parent).to.eq('parent task')
	})
	it('Should not break when the first line is indented', () => {
		const tasks = parseSubtasksViaIndention('  single task', PrefixMode.Default)

		expect(tasks).to.have.length(1)
		expect(tasks[0].parent).toBeNull()
	})
	it('Should add the list of the parent task as list for all sub tasks', () => {
		const tasks = parseSubtasksViaIndention(
`parent task +list
  sub task 1
  sub task 2`, PrefixMode.Default)
		
		expect(tasks).to.have.length(3)
		expect(tasks[0].project).to.eq('list')
		expect(tasks[1].project).to.eq('list')
		expect(tasks[2].project).to.eq('list')
	})
	it('Should clean the indention if there is indention on the first line', () => {
		const tasks = parseSubtasksViaIndention(
`  parent task
  sub task one
    sub task two`, PrefixMode.Default)

		expect(tasks).to.have.length(3)
		expect(tasks[0].parent).toBeNull()
		expect(tasks[0].title).to.eq('parent task')
		expect(tasks[1].title).to.eq('sub task one')
		expect(tasks[1].parent).toBeNull()
		expect(tasks[2].title).to.eq('sub task two')
		expect(tasks[2].parent).to.eq('sub task one')
	})
	it('Should clean the indention if there is indention on the first line but not for subsequent tasks', () => {
		const tasks = parseSubtasksViaIndention(
			`  parent task
  sub task one
first level task one
  sub task two`, PrefixMode.Default)

		expect(tasks).to.have.length(4)
		expect(tasks[0].parent).toBeNull()
		expect(tasks[0].title).to.eq('parent task')
		expect(tasks[1].title).to.eq('sub task one')
		expect(tasks[1].parent).toBeNull()
		expect(tasks[2].title).to.eq('first level task one')
		expect(tasks[2].parent).toBeNull()
		expect(tasks[3].title).to.eq('sub task two')
		expect(tasks[3].parent).to.eq('first level task one')
	})
	it('Should clean the indention if there is indention on the first line for subsequent tasks with less indention', () => {
		const tasks = parseSubtasksViaIndention(
			`  parent task
  sub task one
 first level task one
   sub task two`, PrefixMode.Default)

		expect(tasks).to.have.length(4)
		expect(tasks[0].parent).toBeNull()
		expect(tasks[0].title).to.eq('parent task')
		expect(tasks[1].title).to.eq('sub task one')
		expect(tasks[1].parent).toBeNull()
		expect(tasks[2].title).to.eq('first level task one')
		expect(tasks[2].parent).toBeNull()
		expect(tasks[3].title).to.eq('sub task two')
		expect(tasks[3].parent).to.eq('first level task one')
	})
})
