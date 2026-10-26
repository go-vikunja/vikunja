import {afterAll, afterEach, beforeAll, beforeEach, describe, expect, it, vi} from 'vitest'
import {flushPromises, mount, type VueWrapper} from '@vue/test-utils'
import {createI18n} from 'vue-i18n'
import {reactive, type Directive} from 'vue'
import {readFileSync} from 'node:fs'
import {join} from 'node:path'

const {auth, isMobile} = vi.hoisted(() => ({
	auth: {settings: {timezone: 'UTC', week_start: 1} as {timezone: string, week_start: number}},
	isMobile: {value: false},
}))
vi.mock('@/stores/auth', () => ({useAuthStore: () => auth}))
vi.mock('@/composables/useIsMobile', () => ({useIsMobile: () => isMobile}))
vi.mock('@floating-ui/dom', () => ({
	computePosition: async () => ({x: 0, y: 0}),
	autoUpdate: () => () => {},
	flip: () => ({name: 'flip'}),
	offset: () => ({name: 'offset'}),
	shift: () => ({name: 'shift'}),
	size: () => ({name: 'size'}),
}))

import RepeatAfter from './RepeatAfter.vue'
import type {Task, TaskRepeat} from '@/client/generated'
import type {TaskInput} from '@/client/queries/taskMutations'

// The real English source, so the assertions read the copy people see.
const en = JSON.parse(readFileSync(join(process.cwd(), 'src/i18n/lang/en.json'), 'utf8'))
const i18n = createI18n({legacy: false, locale: 'en', messages: {en}, missingWarn: false, fallbackWarn: false})

// Fri Oct 20 2090, noon UTC: far enough ahead that the preview is shown.
const FRIDAY = '2090-10-20T12:00:00Z'
// Thu Oct 19 2090, 10:30 UTC: "now" for the tasks without a due date.
const THURSDAY_MORNING = new Date('2090-10-19T10:30:00Z')

const MWF: TaskRepeat = {freq: 'weekly', by_day: ['mo', 'we', 'fr']}
const FOREIGN: TaskRepeat = {freq: 'yearly', by_month: [3, 9], by_day: ['1mo']}
const LOCKED: TaskRepeat = {freq: 'yearly', interval: 1, read_only: true}

const PRESETS_FOR_FRIDAY = [
	'Does not repeat',
	'Every day',
	'Every weekday (Mon–Fri)',
	'Every week on Friday',
	'Every month on the 20th',
	'Every month on the 3rd Friday',
	'Every year on Oct 20',
	'Custom…',
]

// The API sends a missing due date in several forms (null, absent, empty, the zero date), so fields stay loose here.
function task(fields: Record<string, unknown> = {}): Task {
	return {id: 1, title: 'Water the plants', due_date: FRIDAY, repeat: null, repeats_from_current_date: false, ...fields} as unknown as Task
}

function deepFreeze<T>(value: T): T {
	if (value !== null && typeof value === 'object' && !Object.isFrozen(value)) {
		Object.freeze(value)
		Object.values(value).forEach(deepFreeze)
	}
	return value
}

// happy-dom has no Popover API: mirror the state the component drives into an attribute.
const POPOVER_OPEN = 'popover-open'

function dispatchToggle(element: HTMLElement, newState: 'open' | 'closed') {
	element.dispatchEvent(Object.assign(new Event('toggle'), {newState}))
}

const noopDirective: Directive = {}

let wrapper: VueWrapper | null = null

async function mountField(modelValue: Task, props: {disabled?: boolean, startOpen?: boolean} = {}) {
	wrapper = mount(RepeatAfter, {
		props: {modelValue: deepFreeze(modelValue), ...props},
		global: {plugins: [i18n], stubs: {Icon: true, RouterLink: true}, directives: {tooltip: noopDirective, cy: noopDirective}},
		attachTo: document.body,
	})
	await flushPromises()
	return wrapper
}

const trigger = () => wrapper!.find<HTMLButtonElement>('.repeat-field-trigger')
const popover = () => document.body.querySelector<HTMLElement>('.popup')
const sheet = () => document.body.querySelector<HTMLDialogElement>('dialog.bottom-sheet')
const confirmDialog = () => document.body.querySelector<HTMLDialogElement>('dialog.modal-dialog:not(.bottom-sheet)')
const options = () => [...document.body.querySelectorAll<HTMLButtonElement>('.repeat-option')]
const option = (label: string) => {
	const found = options().find(button => button.textContent?.trim() === label)
	expect(found, label).toBeDefined()
	return found!
}
const dialogButton = (label: string) => {
	const found = [...confirmDialog()!.querySelectorAll('button')].find(button => button.textContent?.trim() === label)
	expect(found, label).toBeDefined()
	return found!
}
// The accessible-name lookup getByRole('dialog', {name}) performs, for the aria-labelledby / aria-label cases.
function textOf(element: Element, attribute: 'aria-labelledby' | 'aria-describedby'): string {
	return (element.getAttribute(attribute) ?? '').split(/\s+/).filter(Boolean)
		.map(id => document.getElementById(id)?.textContent?.trim() ?? '').join(' ')
}
const customPanel = () => document.body.querySelector<HTMLElement>('.repeat-custom-panel')
const customButton = (label: string) => {
	const found = [...customPanel()!.querySelectorAll('button')].find(button => button.textContent?.trim() === label || button.getAttribute('aria-label') === label)
	expect(found, label).toBeDefined()
	return found!
}
const pressedWeekdays = () => [...document.body.querySelectorAll('.repeat-custom-weekday[aria-pressed="true"]')].map(b => b.getAttribute('aria-label'))
const details = () => wrapper!.find('.repeat-details').text()
const emitted = () => (wrapper!.emitted('update:modelValue') ?? []).map(([value]) => value as TaskInput)
const openEvents = () => (wrapper!.emitted('update:open') ?? []).map(([value]) => value)
const isOpen = () => trigger().attributes('aria-expanded') === 'true'

async function open() {
	trigger().element.focus()
	await trigger().trigger('click')
	await flushPromises()
}

async function click(element: HTMLElement) {
	element.click()
	await flushPromises()
}

// What the browser does to an open popover on Escape or an outside click.
async function lightDismiss() {
	const element = popover()!
	element.removeAttribute(POPOVER_OPEN)
	dispatchToggle(element, 'closed')
	await flushPromises()
}

// Escape on a modal dialog: the browser fires cancel on the topmost one.
async function cancelDialog(dialog: HTMLDialogElement) {
	dialog.dispatchEvent(new Event('cancel', {cancelable: true}))
	await flushPromises()
}

const originalTZ = process.env.TZ
const {showPopover, hidePopover} = HTMLElement.prototype

beforeAll(() => {
	process.env.TZ = 'UTC'
	HTMLElement.prototype.showPopover = function (this: HTMLElement) {
		this.setAttribute(POPOVER_OPEN, '')
		dispatchToggle(this, 'open')
	}
	HTMLElement.prototype.hidePopover = function (this: HTMLElement) {
		this.removeAttribute(POPOVER_OPEN)
		dispatchToggle(this, 'closed')
	}
	vi.spyOn(HTMLDialogElement.prototype, 'showModal').mockImplementation(function (this: HTMLDialogElement) {
		this.setAttribute('open', '')
	})
	vi.spyOn(HTMLDialogElement.prototype, 'close').mockImplementation(function (this: HTMLDialogElement) {
		this.removeAttribute('open')
	})
})

afterAll(() => {
	process.env.TZ = originalTZ
	HTMLElement.prototype.showPopover = showPopover
	HTMLElement.prototype.hidePopover = hidePopover
	vi.restoreAllMocks()
})

beforeEach(() => {
	auth.settings = reactive({timezone: 'UTC', week_start: 1})
	isMobile.value = false
})

afterEach(() => {
	wrapper?.unmount()
	wrapper = null
	document.body.innerHTML = ''
	vi.useRealTimers()
})

describe('RepeatAfter field', () => {
	it('shows the summary with the current due date and the next two dates', async () => {
		await mountField(task({repeat: MWF}))
		expect(trigger().text()).toBe('Every week on Monday, Wednesday, and Friday')
		expect(wrapper!.find('.repeat-preview').element.textContent!.replace(/\s+/g, ' ').trim()).toBe('Next dates: Fri, Oct 20 · Mon, Oct 23 · Wed, Oct 25')
		expect(trigger().attributes('aria-describedby')).toBe(wrapper!.find('.repeat-details').attributes('id'))
	})

	it('says when a task does not repeat', async () => {
		await mountField(task())
		expect(trigger().text()).toBe('Does not repeat')
		expect(wrapper!.find('.repeat-preview').exists()).toBe(false)
	})

	it.each([
		['counted from completion', task({repeat: MWF, repeats_from_current_date: true}), 'Fri, Oct 20'],
		['hourly', task({repeat: {freq: 'hourly'}}), 'Fri, Oct 20'],
		['overdue', task({repeat: MWF, due_date: '2001-10-19T12:00:00Z'}), 'Fri, Oct 19'],
	])('shows only the current due date for a repeat %s', async (_, value, date) => {
		await mountField(value)
		expect(wrapper!.findAll('.repeat-preview').map(p => p.text())).toEqual([`Due ${date}`])
	})

	it('without a timezone setting shows the summary and the hint, and still words presets in the browser zone', async () => {
		auth.settings.timezone = ''
		await mountField(task({repeat: MWF}))
		expect(trigger().text()).toBe('Every week on Monday, Wednesday, and Friday')
		expect(details()).toContain('Set a time zone in your settings to see the next dates.')
		expect(details()).not.toContain('Oct')
		await open()
		expect(options().map(o => o.textContent?.trim())).toEqual(PRESETS_FOR_FRIDAY)
	})

	it('treats a time zone the browser cannot load like an empty one', async () => {
		auth.settings.timezone = 'Not/AZone'
		await mountField(task({repeat: MWF}))
		expect(details()).toContain('Set a time zone in your settings to see the next dates.')
		await open()
		expect(options().map(o => o.textContent?.trim())).toEqual(PRESETS_FOR_FRIDAY)
	})

	it('explains what happens to a repeating task without a due date', async () => {
		await mountField(task({repeat: MWF, due_date: null}))
		expect(details()).toContain('so it will not repeat until one is set')
		await wrapper!.setProps({modelValue: task({repeat: MWF, due_date: null, repeats_from_current_date: true})})
		expect(details()).toContain('it reopens without one')
	})

	it('cannot be opened without write access', async () => {
		await mountField(task({repeat: MWF}), {disabled: true, startOpen: true})
		expect(trigger().attributes('disabled')).toBeDefined()
		expect(isOpen()).toBe(false)
	})
})

describe('presets popover', () => {
	it('opens anchored, with presets worded from the due date, and moves focus into it', async () => {
		await mountField(task())
		expect(trigger().attributes('aria-expanded')).toBe('false')
		expect(trigger().attributes('aria-haspopup')).toBe('dialog')
		await open()
		expect(isOpen()).toBe(true)
		expect(popover()!.hasAttribute(POPOVER_OPEN)).toBe(true)
		const content = document.getElementById(trigger().attributes('aria-controls')!)!
		expect(content.getAttribute('role')).toBe('dialog')
		expect(content.getAttribute('aria-label')).toBe('Repeat')
		expect(popover()!.contains(content)).toBe(true)
		expect(options().map(o => o.textContent?.trim())).toEqual(PRESETS_FOR_FRIDAY)
		expect(document.activeElement).toBe(options()[0])
		expect(openEvents()).toEqual([true])
	})

	it('marks the current preset', async () => {
		await mountField(task({repeat: {freq: 'weekly'}}))
		await open()
		const current = options().filter(o => o.getAttribute('aria-current') === 'true')
		expect(current.map(o => o.textContent?.trim())).toEqual(['Every week on Friday'])
		expect(current[0]!.classList).toContain('currently-active')
	})

	it('marks "Custom…" for an editable repeat that is not a preset', async () => {
		await mountField(task({repeat: MWF}))
		await open()
		expect(options().filter(o => o.getAttribute('aria-current') === 'true').map(o => o.textContent?.trim())).toEqual(['Custom…'])
	})

	it('saves a picked preset once, as a new task object, and closes with focus on the field', async () => {
		const original = task()
		await mountField(original)
		await open()
		await click(option('Every month on the 3rd Friday'))
		expect(emitted()).toHaveLength(1)
		expect(emitted()[0]).not.toBe(original)
		expect(emitted()[0]).toEqual({...original, repeat: {freq: 'monthly', by_day: ['3fr']}})
		expect(isOpen()).toBe(false)
		expect(popover()!.hasAttribute(POPOVER_OPEN)).toBe(false)
		expect(document.activeElement).toBe(trigger().element)
		expect(openEvents()).toEqual([true, false])
	})

	it('removes the repeat with "Does not repeat" by writing null', async () => {
		await mountField(task({repeat: MWF}))
		await open()
		await click(option('Does not repeat'))
		expect(emitted()).toHaveLength(1)
		expect(emitted()[0]).toHaveProperty('repeat', null)
	})

	it('does not save when the current preset is picked again', async () => {
		await mountField(task({repeat: {freq: 'daily'}}))
		await open()
		await click(option('Every day'))
		expect(emitted()).toEqual([])
		expect(isOpen()).toBe(false)
	})

	it('closes on Escape without saving and returns focus to the field', async () => {
		await mountField(task())
		await open()
		await lightDismiss()
		expect(isOpen()).toBe(false)
		expect(emitted()).toEqual([])
		expect(document.activeElement).toBe(trigger().element)
	})

	it('leaves focus where an outside click put it', async () => {
		const elsewhere = document.createElement('input')
		document.body.append(elsewhere)
		await mountField(task())
		await open()
		elsewhere.focus()
		await lightDismiss()
		expect(isOpen()).toBe(false)
		expect(document.activeElement).toBe(elsewhere)
	})

	it('opens on mount when asked to', async () => {
		await mountField(task(), {startOpen: true})
		expect(isOpen()).toBe(true)
		expect(document.activeElement).toBe(options()[0])
		await lightDismiss()
		expect(document.activeElement).toBe(trigger().element)
	})
})

describe('custom form in the popover', () => {
	it('opens pre-filled with the current repeat and focuses its first control', async () => {
		await mountField(task({repeat: MWF}))
		await open()
		await click(option('Custom…'))
		expect(pressedWeekdays()).toEqual(['Monday', 'Wednesday', 'Friday'])
		expect(document.activeElement).toBe(customPanel()!.querySelector('input'))
	})

	it('saves Mon, Wed, Fri with Done', async () => {
		await mountField(task())
		await open()
		await click(option('Custom…'))
		expect(pressedWeekdays()).toEqual(['Friday'])
		await click(customButton('Monday'))
		await click(customButton('Wednesday'))
		await click(customButton('Done'))
		expect(emitted().map(t => t.repeat)).toEqual([MWF])
		expect(isOpen()).toBe(false)
		expect(document.activeElement).toBe(trigger().element)
	})

	it('refuses a count-only raise of the stored repeat', async () => {
		await mountField(task({repeat: {freq: 'weekly', count: 3}}))
		await open()
		await click(option('Custom…'))
		const count = customPanel()!.querySelector<HTMLInputElement>('input[aria-label="Number of times"]')!
		count.value = '5'
		count.dispatchEvent(new Event('input'))
		await flushPromises()
		expect(customPanel()!.textContent).toContain(en.task.repeat.error.countRaise)
		expect(customButton('Done').disabled).toBe(true)
	})

	it('discards an edited draft on Escape, and an unrelated update afterwards saves nothing', async () => {
		await mountField(task({repeat: MWF}))
		await open()
		await click(option('Custom…'))
		await click(customButton('Tuesday'))
		await lightDismiss()
		expect(emitted()).toEqual([])
		await wrapper!.setProps({modelValue: deepFreeze(task({repeat: MWF, title: 'Renamed'}))})
		await flushPromises()
		expect(emitted()).toEqual([])
		await open()
		expect(customPanel()).toBeNull()
		await click(option('Custom…'))
		expect(pressedWeekdays()).toEqual(['Monday', 'Wednesday', 'Friday'])
	})

	it('goes back to the presets without saving and focuses "Custom…"', async () => {
		await mountField(task())
		await open()
		await click(option('Custom…'))
		await click(customButton('Monday'))
		await click(customButton('Back'))
		expect(emitted()).toEqual([])
		expect(isOpen()).toBe(true)
		expect(document.activeElement).toBe(option('Custom…'))
	})
})

describe('a repeat the editor cannot express', () => {
	it('is summarised and marked as not editable here, with only its due date', async () => {
		await mountField(task({repeat: FOREIGN}))
		expect(trigger().text()).toContain('March and September')
		expect(details()).toContain('Set elsewhere: it can\'t be edited here.')
		expect(wrapper!.findAll('.repeat-preview').map(p => p.text())).toEqual(['Due Fri, Oct 20'])
		await open()
		expect(options().filter(o => o.getAttribute('aria-current') === 'true')).toEqual([])
	})

	it('asks before a preset replaces it, then commits once on Replace', async () => {
		await mountField(task({repeat: FOREIGN}))
		await open()
		await click(option('Every day'))
		expect(isOpen()).toBe(false)
		expect(confirmDialog()?.hasAttribute('open')).toBe(true)
		expect(textOf(confirmDialog()!, 'aria-labelledby')).toBe('Replace this repeat?')
		expect(textOf(confirmDialog()!, 'aria-describedby')).toBe(`This task repeats “${trigger().text()}”, which can't be edited here. Replacing it removes that repeat.`)
		expect(openEvents()).toEqual([true])
		await click(confirmDialog()!.querySelector('p')!)
		expect(confirmDialog()).not.toBeNull()
		expect(emitted()).toEqual([])
		await click(dialogButton('Replace'))
		expect(emitted().map(t => t.repeat)).toEqual([{freq: 'daily'}])
		expect(confirmDialog()).toBeNull()
		expect(isOpen()).toBe(false)
		expect(document.activeElement).toBe(trigger().element)
		expect(openEvents()).toEqual([true, false])
	})

	it('keeps the repeat on Cancel and shows the presets again with focus on the chosen one', async () => {
		await mountField(task({repeat: FOREIGN}))
		await open()
		await click(option('Every day'))
		await click(dialogButton('Cancel'))
		expect(confirmDialog()).toBeNull()
		expect(emitted()).toEqual([])
		expect(isOpen()).toBe(true)
		expect(popover()!.hasAttribute(POPOVER_OPEN)).toBe(true)
		expect(document.activeElement).toBe(option('Every day'))
	})

	it('keeps the repeat on Escape in the question and shows the presets again', async () => {
		await mountField(task({repeat: FOREIGN}))
		await open()
		await click(option('Custom…'))
		await cancelDialog(confirmDialog()!)
		expect(confirmDialog()).toBeNull()
		expect(emitted()).toEqual([])
		expect(isOpen()).toBe(true)
		expect(document.activeElement).toBe(option('Custom…'))
	})

	it('opens a fresh Custom form after Replace without writing, and saves on Done', async () => {
		await mountField(task({repeat: FOREIGN}))
		await open()
		await click(option('Custom…'))
		expect(emitted()).toEqual([])
		await click(dialogButton('Replace'))
		expect(emitted()).toEqual([])
		expect(isOpen()).toBe(true)
		expect(pressedWeekdays()).toEqual(['Friday'])
		expect(document.activeElement).toBe(customPanel()!.querySelector('input'))
		await click(customButton('Done'))
		expect(emitted().map(t => t.repeat)).toEqual([{freq: 'weekly'}])
	})

	it('changes only the completion flag', async () => {
		await mountField(task({repeat: FOREIGN}))
		await click(wrapper!.find('.repeat-from-completion input').element as HTMLElement)
		expect(emitted()).toHaveLength(1)
		expect(emitted()[0]!.repeats_from_current_date).toBe(true)
		expect(emitted()[0]!.repeat).toEqual(FOREIGN)
		expect(emitted()[0]!.due_date).toBe(FRIDAY)
	})
})

describe('a repeat the server marks read-only', () => {
	it('shows a generic summary and a note, and offers no presets or Custom', async () => {
		await mountField(task({repeat: LOCKED}))
		expect(trigger().text()).toBe('Repeats on a schedule set elsewhere')
		expect(details()).toContain(en.task.repeat.lockedNote)
		await open()
		expect(isOpen()).toBe(true)
		expect(options()).toEqual([])
		expect(customPanel()).toBeNull()
		const panel = document.body.querySelector('.repeat-locked-panel')!
		expect(panel.textContent).toContain('Repeats on a schedule set elsewhere')
		expect(panel.textContent).toContain(en.task.repeat.lockedNote)
		expect(document.activeElement).toBe(document.getElementById(trigger().attributes('aria-controls')!))
		await lightDismiss()
		expect(emitted()).toEqual([])
		expect(document.activeElement).toBe(trigger().element)
	})

	it('sends the stored rule back unchanged when the completion flag changes', async () => {
		await mountField(task({repeat: LOCKED}))
		await click(wrapper!.find('.repeat-from-completion input').element as HTMLElement)
		expect(emitted()).toHaveLength(1)
		expect(emitted()[0]!.repeat).toEqual(LOCKED)
		expect(emitted()[0]!.repeats_from_current_date).toBe(true)
	})
})

describe('a task without a due date', () => {
	beforeEach(() => {
		vi.useFakeTimers({toFake: ['Date']})
		vi.setSystemTime(THURSDAY_MORNING)
	})

	it('words the presets from today', async () => {
		await mountField(task({due_date: null}))
		await open()
		expect(option('Every week on Thursday')).toBeDefined()
	})

	it.each([
		['null', {due_date: null}],
		['missing', {due_date: undefined}],
		['empty', {due_date: ''}],
		['the API zero date', {due_date: '0001-01-01T00:00:00Z'}],
	])('sets the due date in the same save when it is %s', async (_, fields) => {
		await mountField(task(fields))
		await open()
		await click(option('Every day'))
		expect(emitted()).toHaveLength(1)
		expect(emitted()[0]!.repeat).toEqual({freq: 'daily'})
		expect(emitted()[0]!.due_date).toBe('2090-10-19T12:00:00.000Z')
	})

	it('sets the due date to the first occurrence at the default time and says so', async () => {
		await mountField(task({due_date: null}))
		await open()
		await click(option('Custom…'))
		await click(customButton('Monday'))
		await click(customButton('Thursday'))
		await click(customButton('Done'))
		expect(emitted()).toHaveLength(1)
		expect(emitted()[0]!.repeat).toEqual({freq: 'weekly', by_day: ['mo']})
		expect(emitted()[0]!.due_date).toBe('2090-10-23T12:00:00.000Z')
		await wrapper!.setProps({modelValue: emitted()[0]})
		expect(details()).toContain('The due date was set to Mon, Oct 23 so that the repeat can start.')
	})

	it('still sets the missing due date when the current preset is picked again, in one save', async () => {
		await mountField(task({due_date: null, repeat: {freq: 'daily'}}))
		await open()
		await click(option('Every day'))
		expect(emitted()).toHaveLength(1)
		expect(emitted()[0]!.due_date).toBe('2090-10-19T12:00:00.000Z')
	})

	it('takes today from the settings zone when the browser is still on the day before', async () => {
		auth.settings.timezone = 'Asia/Tokyo'
		// Thu 21:30 in the browser (UTC), already Fri 06:30 in Tokyo.
		vi.setSystemTime(new Date('2090-10-19T21:30:00Z'))
		await mountField(task({due_date: null}))
		await open()
		await click(option('Every week on Friday'))
		// Fri 09:00 in Tokyo, the default time on that day.
		expect(emitted().map(t => t.due_date)).toEqual(['2090-10-20T00:00:00.000Z'])
	})

	it('refuses a repeat that has no date after now', async () => {
		const ended: TaskRepeat = {freq: 'weekly', until: '2090-10-01T23:59:59Z'}
		await mountField(task({due_date: null, repeat: ended}))
		await open()
		await click(option('Custom…'))
		await click(customButton('Done'))
		expect(emitted()).toEqual([])
		expect(wrapper!.find('[role="alert"]').text()).toBe('This repeat has no date after today. Change when it ends and try again.')
		expect(isOpen()).toBe(true)
	})
})

describe('drafts belong to their task', () => {
	it('discards an open Custom draft when the task changes', async () => {
		await mountField(task({repeat: MWF}))
		await open()
		await click(option('Custom…'))
		await click(customButton('Tuesday'))
		await wrapper!.setProps({modelValue: deepFreeze(task({id: 2, repeat: {freq: 'daily'}}))})
		await flushPromises()
		expect(isOpen()).toBe(false)
		expect(customPanel()).toBeNull()
		expect(emitted()).toEqual([])
		expect(openEvents()).toEqual([true, false])
	})

	it('discards an open replace question when the task changes, and its stale button writes nothing', async () => {
		await mountField(task({repeat: FOREIGN}))
		await open()
		await click(option('Every day'))
		const replace = dialogButton('Replace')
		await wrapper!.setProps({modelValue: deepFreeze(task({id: 2, repeat: FOREIGN}))})
		await flushPromises()
		expect(confirmDialog()).toBeNull()
		await click(replace)
		expect(emitted()).toEqual([])
		// The next task's own edit still saves, to that task.
		await open()
		await click(option('Every day'))
		await click(dialogButton('Replace'))
		expect(emitted().map(t => [t.id, t.repeat])).toEqual([[2, {freq: 'daily'}]])
	})

	it('keeps the draft across an update of the same task and saves it to that task once', async () => {
		await mountField(task({repeat: MWF}))
		await open()
		await click(option('Custom…'))
		await click(customButton('Tuesday'))
		await wrapper!.setProps({modelValue: deepFreeze(task({repeat: MWF, title: 'Renamed'}))})
		await flushPromises()
		expect(pressedWeekdays()).toEqual(['Monday', 'Tuesday', 'Wednesday', 'Friday'])
		await click(customButton('Done'))
		expect(emitted()).toHaveLength(1)
		expect(emitted()[0]).toMatchObject({id: 1, title: 'Renamed', repeat: {freq: 'weekly', by_day: ['mo', 'tu', 'we', 'fr']}})
	})
})

describe('on a phone', () => {
	beforeEach(() => {
		isMobile.value = true
	})

	it('opens the presets in a bottom sheet titled Repeat and saves a pick', async () => {
		await mountField(task())
		await open()
		expect(popover()).toBeNull()
		expect(sheet()?.getAttribute('aria-label')).toBe('Repeat')
		expect(sheet()!.contains(options()[0]!)).toBe(true)
		// The sheet dialog carries the name, so the content inside it is not a second dialog.
		expect(document.getElementById(trigger().attributes('aria-controls')!)!.hasAttribute('role')).toBe(false)
		expect(options().map(o => o.textContent?.trim())).toEqual(PRESETS_FOR_FRIDAY)
		await click(option('Every day'))
		expect(emitted().map(t => t.repeat)).toEqual([{freq: 'daily'}])
		expect(sheet()).toBeNull()
		expect(document.activeElement).toBe(trigger().element)
	})

	it('closes the sheet on Escape without saving', async () => {
		await mountField(task())
		await open()
		await cancelDialog(sheet()!)
		expect(sheet()).toBeNull()
		expect(isOpen()).toBe(false)
		expect(emitted()).toEqual([])
	})

	it('stacks the replace question on the sheet, and Escape closes only the question', async () => {
		await mountField(task({repeat: FOREIGN}))
		await open()
		await click(option('Every day'))
		expect(sheet()?.hasAttribute('open')).toBe(true)
		expect(confirmDialog()?.hasAttribute('open')).toBe(true)
		await cancelDialog(confirmDialog()!)
		expect(confirmDialog()).toBeNull()
		expect(sheet()).not.toBeNull()
		expect(emitted()).toEqual([])
		expect(document.activeElement).toBe(option('Every day'))
		await click(option('Custom…'))
		await click(dialogButton('Replace'))
		expect(sheet()!.contains(customPanel()!)).toBe(true)
		expect(emitted()).toEqual([])
		await click(customButton('Done'))
		expect(emitted().map(t => t.repeat)).toEqual([{freq: 'weekly'}])
		expect(sheet()).toBeNull()
	})
})
