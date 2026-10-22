import {test, expect} from '../../support/fixtures'
import {TaskFactory} from '../../factories/task'
import {ProjectFactory} from '../../factories/project'

const NESTED_CHECKLIST = `<ul data-type="taskList">
	<li data-checked="true" data-type="taskItem"><label><input type="checkbox" checked><span></span></label>
		<div>
			<p>Parent item</p>
			<ul data-type="taskList">
				<li data-checked="false" data-type="taskItem"><label><input type="checkbox"><span></span></label>
					<div><p>Child item</p></div>
				</li>
			</ul>
		</div>
	</li>
</ul>`

test.describe('Nested checklist items', () => {
	test.beforeEach(async () => {
		await ProjectFactory.create(1)
	})

	/**
	 * Regression test for https://github.com/go-vikunja/vikunja/issues/3712
	 *
	 * Styling the whole `li[data-checked=true]` also hit the nested task list living
	 * inside the item's content div, so an unchecked child rendered greyed out and
	 * struck through - it looked checked.
	 */
	test('Should not strike through an unchecked child of a checked item (issue #3712)', async ({authenticatedPage: page}) => {
		const tasks = await TaskFactory.create(1, {
			id: 1,
			description: NESTED_CHECKLIST,
		})

		await page.goto(`/tasks/${tasks[0].id}`)

		const description = page.locator('.task-view .details.content.description .tiptap')
		const checkedItem = description.locator('li[data-checked="true"] > div > p').first()
		const uncheckedItem = description.locator('li[data-checked="false"] > div > p').first()

		await expect(checkedItem).toHaveText('Parent item')
		await expect(uncheckedItem).toHaveText('Child item')

		// text-decoration propagates to descendants and cannot be reset by a child, so the
		// only reliable check is that no ancestor of the child carries the strikethrough.
		const struckAncestors = await uncheckedItem.evaluate(el => {
			const decorated: string[] = []
			for (let node = el; node; node = node.parentElement) {
				if (node.classList.contains('tiptap')) {
					break
				}
				if (getComputedStyle(node).textDecorationLine.includes('line-through')) {
					decorated.push(node.tagName)
				}
			}
			return decorated
		})
		expect(struckAncestors).toEqual([])

		// The checked parent still gets its own strikethrough and muted colour.
		await expect(checkedItem).toHaveCSS('text-decoration-line', 'line-through')
		expect(await uncheckedItem.evaluate(el => getComputedStyle(el).color))
			.not.toBe(await checkedItem.evaluate(el => getComputedStyle(el).color))
	})
})
