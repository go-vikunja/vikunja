import {afterAll, afterEach, beforeAll, beforeEach, describe, expect, it, vi} from 'vitest'
import {flushPromises, mount, type VueWrapper} from '@vue/test-utils'
import {createI18n} from 'vue-i18n'
import {QueryClient, VueQueryPlugin} from '@tanstack/vue-query'
import {reactive, type Directive} from 'vue'
import {readFileSync} from 'node:fs'
import {join} from 'node:path'

const {auth, isMobile} = vi.hoisted(() => ({
	auth: {settings: {timezone: 'UTC', week_start: 1} as {timezone: string, week_start: number}},
	isMobile: {value: false},
}))
vi.mock('@/stores/auth', () => ({useAuthStore: () => auth}))
const sdk = vi.hoisted(() => ({
	tasksOccurrencesPreview: vi.fn(),
}))
vi.mock('@/client/generated', () => sdk)
vi.mock('@/composables/useIsMobile', () => ({useIsMobile: () => isMobile}))
vi.mock('@floating-ui/dom', () => ({
	computePosition: async () => ({x: 0, y: 0}),
	autoUpdate: () => () => {},
	flip: () => ({name: 'flip'}),
	offset: () => ({name: 'offset'}),
	shift: () => ({name: 'shift'}),
	size: () => ({name: 'size'}),
}))

import RepeatAfter, {type RepeatSaveOptions, type RepeatSaveResult} from './RepeatAfter.vue'
import type {Task, TaskRepeat} from '@/client/generated'
import type {TaskInput, WriteOutcome} from '@/client/queries/taskMutations'
import {isRefusal, previewFixtures} from '@/client/queries/occurrencePreview.fixtures'
import type {PreviewAnswer} from '@/client/generated'
import {PREVIEW_DEBOUNCE_MS} from '@/composables/useOccurrencePreview'

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

// Browsers fire beforetoggle while the state changes and toggle afterwards.
function dispatchToggle(element: HTMLElement, newState: 'open' | 'closed') {
	element.dispatchEvent(Object.assign(new Event('beforetoggle'), {newState}))
	element.dispatchEvent(Object.assign(new Event('toggle'), {newState}))
}

const noopDirective: Directive = {}

function answerOf(row: string): PreviewAnswer {
	const answer = previewFixtures.find(fixture => fixture.row === row)?.answer
	if (!answer || isRefusal(answer)) throw new Error(`no answer for ${row}`)
	return answer
}

let wrapper: VueWrapper | null = null
let queryClient: QueryClient
const save = vi.fn<(task: TaskInput, options: RepeatSaveOptions) => Promise<RepeatSaveResult>>()

async function mountField(modelValue: Task, props: {disabled?: boolean, startOpen?: boolean, writeOutcome?: WriteOutcome | null} = {}) {
	wrapper = mount(RepeatAfter, {
		props: {modelValue: deepFreeze(modelValue), save, ...props},
		global: {plugins: [i18n, [VueQueryPlugin, {queryClient}]], stubs: {Icon: true, RouterLink: true}, directives: {tooltip: noopDirective, cy: noopDirective}},
		attachTo: document.body,
	})
	await settle()
	return wrapper
}

// Lets the preview debounce, the popup's own timers and every pending answer run, as a person pausing would.
async function settle() {
	await vi.advanceTimersByTimeAsync(PREVIEW_DEBOUNCE_MS)
	await flushPromises()
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
const previewText = () => wrapper!.find('.repeat-preview').element.textContent!.replace(/\s+/g, ' ').trim()
const previewRows = () => wrapper!.findAll('.repeat-preview-date').map(row => row.text())
const previewRequests = () => sdk.tasksOccurrencesPreview.mock.calls.map(([options]) => (options as {body: Record<string, unknown>}).body)
const saves = () => save.mock.calls.map(([value]) => value as TaskInput)
const saveOptions = () => save.mock.calls.map(([, options]) => options)
const openEvents = () => (wrapper!.emitted('update:open') ?? []).map(([value]) => value)
const isOpen = () => trigger().attributes('aria-expanded') === 'true'

async function open() {
	trigger().element.focus()
	await trigger().trigger('click')
	await settle()
}

async function click(element: HTMLElement) {
	element.click()
	await settle()
}

// What the browser does to an open popover on Escape or an outside click.
async function lightDismiss() {
	const element = popover()!
	element.removeAttribute(POPOVER_OPEN)
	dispatchToggle(element, 'closed')
	await settle()
}

// Escape on a modal dialog: the browser fires cancel on the topmost one.
async function cancelDialog(dialog: HTMLDialogElement) {
	dialog.dispatchEvent(new Event('cancel', {cancelable: true}))
	await settle()
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
	vi.useFakeTimers({toFake: ['setTimeout', 'clearTimeout', 'Date']})
	vi.setSystemTime(THURSDAY_MORNING)
	queryClient = new QueryClient()
	sdk.tasksOccurrencesPreview.mockReset()
	sdk.tasksOccurrencesPreview.mockImplementation(async () => ({data: answerOf('P-10')}))
	save.mockReset()
	save.mockResolvedValue({ok: true, outcome: {}})
})

afterEach(() => {
	wrapper?.unmount()
	wrapper = null
	document.body.innerHTML = ''
	queryClient.clear()
	vi.useRealTimers()
})

describe('RepeatAfter field', () => {
	it('shows the due date and the server\'s next dates in the server zone', async () => {
		await mountField(task({repeat: MWF}))
		expect(trigger().text()).toBe('Every week on Monday, Wednesday, and Friday')
		expect(previewText()).toBe('Next dates: Thu, Oct 29 · Mon, Nov 2 · Thu, Nov 5 · Mon, Nov 9 · Thu, Nov 12')
		expect(previewRequests()).toEqual([{repeat: MWF, due_date: new Date(FRIDAY).toISOString(), task_id: 1, repeats_from_current_date: false, count: 10}])
		expect(trigger().attributes('aria-describedby')).toBe(wrapper!.find('.repeat-details').attributes('id'))
	})

	it.each([
		[[28, 29, 30, 31], 'Every month on the last day'],
		[[28, 29, 30], 'Every month on the 30th, or the last day of February'],
	])('summarises a migrated month-end rule (BYMONTHDAY=%j;BYSETPOS=-1) in words', async (days, summary) => {
		await mountField(task({repeat: {freq: 'monthly', interval: 1, by_month_day: days, by_set_pos: [-1]}}))
		expect(trigger().text()).toBe(summary)
	})

	it('says when a task does not repeat', async () => {
		await mountField(task())
		expect(trigger().text()).toBe('Does not repeat')
		expect(wrapper!.find('.repeat-preview').exists()).toBe(false)
		expect(previewRequests()).toEqual([])
	})

	it.each([
		['counted from completion', task({repeat: MWF, repeats_from_current_date: true})],
		['hourly', task({repeat: {freq: 'hourly'}})],
		['overdue', task({repeat: MWF, due_date: '2001-10-19T12:00:00Z'})],
	])('previews hourly, overdue and from-completion rules from the server: %s', async (_, value) => {
		await mountField(value)
		expect(previewRequests()).toHaveLength(1)
		expect(previewRows()).toHaveLength(5)
	})

	it('labels a from-completion preview as an on-time forecast', async () => {
		sdk.tasksOccurrencesPreview.mockResolvedValue({data: answerOf('P-13')})
		await mountField(task({repeat: {freq: 'hourly'}, repeats_from_current_date: true}))
		expect(previewRows()).toEqual(['Oct 28, 6:00 PM ·', 'Oct 28, 7:00 PM ·', 'Oct 28, 8:00 PM ·', 'Oct 28, 9:00 PM'])
		expect(details()).toContain(en.task.repeat.preview.forecast)
	})

	it('shows a time the clocks pass twice once per instant, told apart by its offset', async () => {
		sdk.tasksOccurrencesPreview.mockResolvedValue({data: {
			due_date: '2026-10-25T01:30:00+02:00',
			due_date_source: 'explicit',
			occurrences: ['2026-10-25T02:30:00+02:00', '2026-10-25T02:30:00+01:00', '2026-10-25T03:30:00+01:00'],
			timezone: 'Europe/Berlin',
			evaluated_at: '2026-10-24T10:00:00+02:00',
		}})
		await mountField(task({repeat: {freq: 'hourly'}, due_date: '2026-10-24T23:30:00Z'}))
		expect(previewRows()).toEqual(['Oct 25, 1:30 AM ·', 'Oct 25, 2:30 AM (GMT+2) ·', 'Oct 25, 2:30 AM (GMT+1) ·', 'Oct 25, 3:30 AM'])
	})

	it('formats the server answer in the zone it names, whatever the settings', async () => {
		// Fri 01:00 in Pago Pago, where 09:00 in Berlin is still the day before.
		auth.settings.timezone = 'Pacific/Pago_Pago'
		await mountField(task({repeat: MWF}))
		expect(previewRows()[0]).toBe('Thu, Oct 29 ·')
		await open()
		expect(options().map(o => o.textContent?.trim())).toEqual(PRESETS_FOR_FRIDAY)
	})

	it('marks the last answer as updating while a new one loads, in a polite live region', async () => {
		await mountField(task({repeat: MWF}))
		sdk.tasksOccurrencesPreview.mockImplementation(() => new Promise(() => {}))
		await wrapper!.setProps({modelValue: deepFreeze(task({repeat: {freq: 'weekly', by_day: ['mo']}}))})
		await settle()
		const region = wrapper!.find('.repeat-details [aria-live="polite"]')
		expect(region.find('.repeat-preview-updating').text()).toBe(en.task.repeat.preview.updating)
		expect(region.findAll('.repeat-preview-date')).toHaveLength(5)
	})

	it('says the preview is unavailable when it cannot load, in a polite live region', async () => {
		sdk.tasksOccurrencesPreview.mockRejectedValue(new TypeError('Failed to fetch'))
		await mountField(task({repeat: MWF}))
		expect(wrapper!.find('.repeat-details [aria-live="polite"] .repeat-preview-unavailable').text()).toBe(en.task.repeat.preview.unavailable)
		expect(trigger().attributes('disabled')).toBeUndefined()
	})

	it('shows no preview for a locked rule or an undated done task', async () => {
		await mountField(task({repeat: LOCKED}))
		expect(details()).toContain(en.task.repeat.lockedNote)
		wrapper!.unmount()
		await mountField(task({repeat: MWF, due_date: null, done: true}))
		await open()
		expect(previewRequests()).toEqual([])
		expect(details()).toContain('so it will not repeat until one is set')
	})

	it('explains what happens to a repeating task without a due date, and asks for nothing while closed', async () => {
		await mountField(task({repeat: MWF, due_date: null}))
		expect(details()).toContain('so it will not repeat until one is set')
		await wrapper!.setProps({modelValue: task({repeat: MWF, due_date: null, repeats_from_current_date: true})})
		expect(details()).toContain('it reopens without one')
		await settle()
		expect(previewRequests()).toEqual([])
	})

	it('previews the date a save would get once the editor is open on a task without a due date', async () => {
		sdk.tasksOccurrencesPreview.mockResolvedValue({data: answerOf('P-1')})
		await mountField(task({repeat: {freq: 'daily'}, due_date: null}))
		await open()
		expect(previewRequests()).toEqual([{repeat: {freq: 'daily'}, initial_due_time: '12:00', task_id: 1, repeats_from_current_date: false, count: 10}])
		expect(previewText()).toBe('Next dates: Thu, Oct 29 · Fri, Oct 30 · Sat, Oct 31')
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
		expect(saves()).toHaveLength(1)
		expect(saves()[0]).not.toBe(original)
		expect(saves()[0]).toEqual({...original, repeat: {freq: 'monthly', by_day: ['3fr']}})
		expect(isOpen()).toBe(false)
		expect(popover()!.hasAttribute(POPOVER_OPEN)).toBe(false)
		expect(document.activeElement).toBe(trigger().element)
		expect(openEvents()).toEqual([true, false])
	})

	it('removes the repeat with "Does not repeat" by writing null', async () => {
		await mountField(task({repeat: MWF}))
		await open()
		await click(option('Does not repeat'))
		expect(saves()).toHaveLength(1)
		expect(saves()[0]).toHaveProperty('repeat', null)
	})

	it('does not save when the current preset is picked again', async () => {
		await mountField(task({repeat: {freq: 'daily'}}))
		await open()
		await click(option('Every day'))
		expect(saves()).toEqual([])
		expect(isOpen()).toBe(false)
	})

	it('closes on Escape without saving and returns focus to the field', async () => {
		await mountField(task())
		await open()
		await lightDismiss()
		expect(isOpen()).toBe(false)
		expect(saves()).toEqual([])
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
		expect(saves().map(t => t.repeat)).toEqual([MWF])
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
		await settle()
		expect(customPanel()!.textContent).toContain(en.task.repeat.error.countRaise)
		expect(customButton('Done').disabled).toBe(true)
	})

	it('discards an edited draft on Escape, and an unrelated update afterwards saves nothing', async () => {
		await mountField(task({repeat: MWF}))
		await open()
		await click(option('Custom…'))
		await click(customButton('Tuesday'))
		await lightDismiss()
		expect(saves()).toEqual([])
		await wrapper!.setProps({modelValue: deepFreeze(task({repeat: MWF, title: 'Renamed'}))})
		await settle()
		expect(saves()).toEqual([])
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
		expect(saves()).toEqual([])
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
		expect(saves()).toEqual([])
		await click(dialogButton('Replace'))
		expect(saves().map(t => t.repeat)).toEqual([{freq: 'daily'}])
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
		expect(saves()).toEqual([])
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
		expect(saves()).toEqual([])
		expect(isOpen()).toBe(true)
		expect(document.activeElement).toBe(option('Custom…'))
	})

	it('opens a fresh Custom form after Replace without writing, and saves on Done', async () => {
		await mountField(task({repeat: FOREIGN}))
		await open()
		await click(option('Custom…'))
		expect(saves()).toEqual([])
		await click(dialogButton('Replace'))
		expect(saves()).toEqual([])
		expect(isOpen()).toBe(true)
		expect(pressedWeekdays()).toEqual(['Friday'])
		expect(document.activeElement).toBe(customPanel()!.querySelector('input'))
		await click(customButton('Done'))
		expect(saves().map(t => t.repeat)).toEqual([{freq: 'weekly'}])
	})

	it('changes only the completion flag', async () => {
		await mountField(task({repeat: FOREIGN}))
		await click(wrapper!.find('.repeat-from-completion input').element as HTMLElement)
		expect(saves()).toHaveLength(1)
		expect(saves()[0]!.repeats_from_current_date).toBe(true)
		expect(saves()[0]!.repeat).toEqual(FOREIGN)
		expect(saves()[0]!.due_date).toBe(FRIDAY)
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
		expect(saves()).toEqual([])
		expect(document.activeElement).toBe(trigger().element)
	})

	it('sends the stored rule back unchanged when the completion flag changes', async () => {
		await mountField(task({repeat: LOCKED}))
		await click(wrapper!.find('.repeat-from-completion input').element as HTMLElement)
		expect(saves()).toHaveLength(1)
		expect(saves()[0]!.repeat).toEqual(LOCKED)
		expect(saves()[0]!.repeats_from_current_date).toBe(true)
	})
})

describe('a task without a due date', () => {
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
	])('sends the time fixed at opening instead of a due date when it is %s', async (_, fields) => {
		await mountField(task(fields))
		await open()
		await click(option('Every day'))
		expect(saves()).toHaveLength(1)
		expect(saves()[0]!.repeat).toEqual({freq: 'daily'})
		expect(saves()[0]!.due_date).toBe(task(fields).due_date)
		expect(saveOptions()).toEqual([{initialDueTime: '12:00'}])
	})

	it('saves an undated task with the time fixed at open, shows saving, then the server\'s date and notice', async () => {
		let land: (result: RepeatSaveResult) => void = () => {}
		save.mockImplementation(() => new Promise(resolve => {
			land = resolve
		}))
		await mountField(task({due_date: null}))
		await open()
		await click(option('Custom…'))
		await click(customButton('Monday'))
		await click(customButton('Thursday'))
		await click(customButton('Done'))

		expect(saves().map(t => [t.repeat, t.due_date])).toEqual([[{freq: 'weekly', by_day: ['mo']}, null]])
		expect(saveOptions()).toEqual([{initialDueTime: '12:00'}])
		expect(isOpen()).toBe(true)
		expect(document.body.querySelector('.repeat-popup [aria-live="polite"].repeat-commit-status')!.textContent!.trim()).toBe(en.task.repeat.saving)
		expect(customButton('Done').disabled).toBe(true)
		expect(details()).not.toContain('The due date was set')

		land({ok: true, outcome: {dueDateSource: 'generated'}})
		await settle()
		await wrapper!.setProps({
			modelValue: deepFreeze(task({due_date: '2090-10-23T12:00:00Z', repeat: {freq: 'weekly', by_day: ['mo']}})),
			writeOutcome: {dueDateSource: 'generated'},
		})

		expect(isOpen()).toBe(false)
		expect(details()).toContain('The due date was set to Mon, Oct 23 so that the repeat can start.')
	})

	it('a re-picked preset on an undated task sends one seed request', async () => {
		await mountField(task({due_date: null, repeat: {freq: 'daily'}}))
		await open()
		await click(option('Every day'))
		expect(saves()).toHaveLength(1)
		expect(saveOptions()).toEqual([{initialDueTime: '12:00'}])
	})

	it('the fixed time is sent unchanged; the server\'s date is shown', async () => {
		auth.settings.timezone = 'Asia/Tokyo'
		// Thu 21:30 in the browser (UTC), already Fri 06:30 in Tokyo.
		vi.setSystemTime(new Date('2090-10-19T21:30:00Z'))
		await mountField(task({due_date: null}))
		await open()
		// Past the next default hour, but the time was fixed when the editor opened.
		vi.setSystemTime(new Date('2090-10-20T03:30:00Z'))
		await click(option('Every week on Friday'))
		expect(saveOptions()).toEqual([{initialDueTime: '09:00'}])
		expect(saves()[0]!.due_date).toBeNull()

		await wrapper!.setProps({
			modelValue: deepFreeze(task({due_date: '2090-10-27T00:00:00Z', repeat: {freq: 'weekly'}})),
			writeOutcome: {dueDateSource: 'generated'},
		})
		expect(details()).toContain('The due date was set to Fri, Oct 27 so that the repeat can start.')
	})

	it('saves undated and shows the server\'s reason', async () => {
		const ended: TaskRepeat = {freq: 'weekly', until: '2090-10-01T23:59:59Z'}
		save.mockResolvedValue({ok: true, outcome: {reason: 'until_passed'}})
		await mountField(task({due_date: null, repeat: ended}))
		await open()
		await click(option('Custom…'))
		await click(customButton('Done'))
		expect(saveOptions()).toEqual([{initialDueTime: '12:00'}])
		expect(isOpen()).toBe(false)

		await wrapper!.setProps({writeOutcome: {reason: 'until_passed'}})
		expect(details()).toContain(en.task.repeat.reason.until_passed)
		expect(details()).not.toContain('The due date was set')
	})

	it('never asks the server for a due date for a done task', async () => {
		await mountField(task({due_date: null, done: true}))
		await open()
		await click(option('Every day'))
		expect(saveOptions()).toEqual([{}])
	})
})

describe('saving', () => {
	it('disables the choices while a save is in flight and accepts no second one', async () => {
		save.mockImplementation(() => new Promise(() => {}))
		await mountField(task())
		await open()
		await click(option('Every day'))
		expect(option('Every week on Friday').getAttribute('aria-disabled')).toBe('true')
		expect(wrapper!.find<HTMLInputElement>('.repeat-from-completion input').element.disabled).toBe(true)
		await click(option('Every week on Friday'))
		expect(saves()).toHaveLength(1)
	})

	it('announces saving politely and leaves no control in the popup without a name', async () => {
		save.mockImplementation(() => new Promise(() => {}))
		await mountField(task({repeat: MWF}))
		await open()
		await click(option('Custom…'))
		await click(customButton('Done'))

		const status = document.body.querySelector('.repeat-popup .repeat-commit-status')!
		expect(status.getAttribute('aria-live')).toBe('polite')
		expect(status.textContent!.trim()).toBe(en.task.repeat.saving)
		const unnamed = [...document.body.querySelectorAll('.repeat-popup button, .repeat-popup input, .repeat-popup select')]
			.filter(control => !control.getAttribute('aria-label') && !control.textContent?.trim() && !control.closest('label')?.textContent?.trim()
				&& !(control.id && document.querySelector(`label[for="${control.id}"]`)))
			.map(control => control.outerHTML)
		expect(unnamed).toEqual([])
	})

	it('keeps the draft open with an error when the save fails, and saves it on the next try', async () => {
		save.mockResolvedValueOnce({ok: false, error: new Error('nope')})
		await mountField(task({repeat: MWF}))
		await open()
		await click(option('Custom…'))
		await click(customButton('Tuesday'))
		await click(customButton('Done'))

		expect(isOpen()).toBe(true)
		expect(pressedWeekdays()).toEqual(['Monday', 'Tuesday', 'Wednesday', 'Friday'])
		expect(document.body.querySelector('.repeat-popup [role="alert"]')!.textContent!.trim()).toBe(en.task.repeat.error.saveFailed)

		await click(customButton('Done'))
		expect(saves().map(t => t.repeat)).toEqual([
			{freq: 'weekly', by_day: ['mo', 'tu', 'we', 'fr']},
			{freq: 'weekly', by_day: ['mo', 'tu', 'we', 'fr']},
		])
		expect(isOpen()).toBe(false)
	})

	it('refuses to save while offline and keeps the draft', async () => {
		const onLine = vi.spyOn(navigator, 'onLine', 'get').mockReturnValue(false)
		await mountField(task({repeat: MWF}))
		window.dispatchEvent(new Event('offline'))
		await open()
		await click(option('Custom…'))
		await click(customButton('Tuesday'))
		await click(customButton('Done'))
		onLine.mockRestore()

		expect(saves()).toEqual([])
		expect(isOpen()).toBe(true)
		expect(pressedWeekdays()).toEqual(['Monday', 'Tuesday', 'Wednesday', 'Friday'])
		expect(document.body.querySelector('.repeat-popup [role="alert"]')!.textContent!.trim()).toBe(en.task.repeat.error.offline)
	})

	describe('a save that is still running', () => {
		let fail: () => void = () => {}

		beforeEach(() => {
			save.mockImplementation(() => new Promise(resolve => {
				fail = () => resolve({ok: false, error: new Error('nope')})
			}))
		})

		async function startSavingDraft() {
			await mountField(task({repeat: MWF}))
			await open()
			await click(option('Custom…'))
			await click(customButton('Tuesday'))
			await click(customButton('Done'))
			expect(saves()).toHaveLength(1)
		}

		async function expectDraftKeptWithError() {
			fail()
			await settle()
			expect(isOpen()).toBe(true)
			expect(pressedWeekdays()).toEqual(['Monday', 'Tuesday', 'Wednesday', 'Friday'])
			expect(document.body.querySelector('.repeat-popup [role="alert"]')!.textContent!.trim()).toBe(en.task.repeat.error.saveFailed)
		}

		it.each([
			['Back', () => click(customButton('Back'))],
			['the field button', () => trigger().trigger('click').then(settle)],
			['Escape', lightDismiss],
			['an outside click', async () => {
				const elsewhere = document.createElement('input')
				document.body.append(elsewhere)
				elsewhere.focus()
				await lightDismiss()
			}],
		])('keeps the draft open through %s until the save fails', async (_, act) => {
			await startSavingDraft()
			await act()
			expect(isOpen()).toBe(true)
			expect(popover()!.hasAttribute(POPOVER_OPEN)).toBe(true)
			expect(pressedWeekdays()).toEqual(['Monday', 'Tuesday', 'Wednesday', 'Friday'])
			await expectDraftKeptWithError()
		})

		it('keeps the sheet and the draft on a phone through Escape until the save fails', async () => {
			isMobile.value = true
			await startSavingDraft()
			await cancelDialog(sheet()!)
			expect(sheet()?.hasAttribute('open')).toBe(true)
			expect(pressedWeekdays()).toEqual(['Monday', 'Tuesday', 'Wednesday', 'Friday'])
			await expectDraftKeptWithError()
		})

		it('keeps the replace question through Escape and Cancel until the save fails', async () => {
			await mountField(task({repeat: FOREIGN}))
			await open()
			await click(option('Every day'))
			await click(dialogButton('Replace'))
			await cancelDialog(confirmDialog()!)
			expect(confirmDialog()?.hasAttribute('open')).toBe(true)
			await click(dialogButton('Cancel'))
			expect(confirmDialog()?.hasAttribute('open')).toBe(true)

			fail()
			await settle()
			expect(confirmDialog()).toBeNull()
			expect(isOpen()).toBe(true)
			expect(document.body.querySelector('.repeat-popup [role="alert"]')!.textContent!.trim()).toBe(en.task.repeat.error.saveFailed)
			expect(saves().map(t => t.repeat)).toEqual([{freq: 'daily'}])
		})
	})

	it('retries a failed replace with the time fixed when the editor opened, past a default-hour boundary', async () => {
		let fail: () => void = () => {}
		save.mockImplementationOnce(() => new Promise(resolve => {
			fail = () => resolve({ok: false, error: new Error('nope')})
		}))
		await mountField(task({repeat: FOREIGN, due_date: null}))
		await open()
		await click(option('Every day'))
		await click(dialogButton('Replace'))
		// 12:30 would pick 15:00, but the editing session started at 10:30.
		vi.setSystemTime(new Date('2090-10-19T12:30:00Z'))
		fail()
		await settle()
		expect(isOpen()).toBe(true)

		await click(option('Every day'))
		await click(dialogButton('Replace'))

		expect(saveOptions()).toEqual([{initialDueTime: '12:00'}, {initialDueTime: '12:00'}])
		expect(isOpen()).toBe(false)
	})

	it('shows a failed completion-flag change beside the field', async () => {
		save.mockResolvedValue({ok: false, error: new Error('nope')})
		await mountField(task({repeat: MWF}))
		await click(wrapper!.find('.repeat-from-completion input').element as HTMLElement)
		expect(saveOptions()).toEqual([{}])
		expect(wrapper!.find('.repeat-details [role="alert"]').text()).toBe(en.task.repeat.error.saveFailed)
	})

	it('drops a preview answer for the rule a save replaced', async () => {
		const answers: ((answer: PreviewAnswer) => void)[] = []
		sdk.tasksOccurrencesPreview.mockImplementation(() => new Promise(resolve => {
			answers.push(answer => resolve({data: answer}))
		}))
		await mountField(task({repeat: MWF}))
		expect(answers).toHaveLength(1)

		// The save landed: the parent shows the saved task and drops what was asked about the old one.
		await wrapper!.setProps({modelValue: deepFreeze(task({repeat: {freq: 'yearly'}}))})
		await (wrapper!.vm as unknown as {cancelPreview: () => Promise<void>}).cancelPreview()
		answers[0]!(answerOf('P-10'))
		await flushPromises()
		expect(previewRows()).toEqual([])

		await settle()
		answers[1]!(answerOf('matrix: yearly implicit'))
		await flushPromises()
		expect(previewRows()).toEqual(['Wed, Mar 14 ·', 'Fri, Mar 14 ·', 'Sat, Mar 14'])
	})
})

describe('a task without a due date across midnight', () => {
	it('asks to pick again instead of saving when the day changed since the editor opened', async () => {
		vi.setSystemTime(new Date('2090-10-19T23:59:00Z'))
		await mountField(task({due_date: null}))
		await open()
		expect(option('Every week on Thursday')).toBeDefined()
		vi.setSystemTime(new Date('2090-10-20T00:01:00Z'))

		await click(option('Every day'))

		expect(saves()).toEqual([])
		expect(isOpen()).toBe(true)
		expect(wrapper!.find('[role="alert"]').text()).toBe(en.task.repeat.error.dayChanged)
		expect(option('Every week on Friday')).toBeDefined()
		await click(option('Every day'))
		expect(saves().map(t => t.repeat)).toEqual([{freq: 'daily'}])
	})
})

describe('drafts belong to their task', () => {
	it('discards an open Custom draft when the task changes', async () => {
		await mountField(task({repeat: MWF}))
		await open()
		await click(option('Custom…'))
		await click(customButton('Tuesday'))
		await wrapper!.setProps({modelValue: deepFreeze(task({id: 2, repeat: {freq: 'daily'}}))})
		await settle()
		expect(isOpen()).toBe(false)
		expect(customPanel()).toBeNull()
		expect(saves()).toEqual([])
		expect(openEvents()).toEqual([true, false])
	})

	it('discards an open replace question when the task changes, and its stale button writes nothing', async () => {
		await mountField(task({repeat: FOREIGN}))
		await open()
		await click(option('Every day'))
		const replace = dialogButton('Replace')
		await wrapper!.setProps({modelValue: deepFreeze(task({id: 2, repeat: FOREIGN}))})
		await settle()
		expect(confirmDialog()).toBeNull()
		await click(replace)
		expect(saves()).toEqual([])
		// The next task's own edit still saves, to that task.
		await open()
		await click(option('Every day'))
		await click(dialogButton('Replace'))
		expect(saves().map(t => [t.id, t.repeat])).toEqual([[2, {freq: 'daily'}]])
	})

	it('keeps the draft across an update of the same task and saves it to that task once', async () => {
		await mountField(task({repeat: MWF}))
		await open()
		await click(option('Custom…'))
		await click(customButton('Tuesday'))
		await wrapper!.setProps({modelValue: deepFreeze(task({repeat: MWF, title: 'Renamed'}))})
		await settle()
		expect(pressedWeekdays()).toEqual(['Monday', 'Tuesday', 'Wednesday', 'Friday'])
		await click(customButton('Done'))
		expect(saves()).toHaveLength(1)
		expect(saves()[0]).toMatchObject({id: 1, title: 'Renamed', repeat: {freq: 'weekly', by_day: ['mo', 'tu', 'we', 'fr']}})
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
		expect(saves().map(t => t.repeat)).toEqual([{freq: 'daily'}])
		expect(sheet()).toBeNull()
		expect(document.activeElement).toBe(trigger().element)
	})

	it('closes the sheet on Escape without saving', async () => {
		await mountField(task())
		await open()
		await cancelDialog(sheet()!)
		expect(sheet()).toBeNull()
		expect(isOpen()).toBe(false)
		expect(saves()).toEqual([])
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
		expect(saves()).toEqual([])
		expect(document.activeElement).toBe(option('Every day'))
		await click(option('Custom…'))
		await click(dialogButton('Replace'))
		expect(sheet()!.contains(customPanel()!)).toBe(true)
		expect(saves()).toEqual([])
		await click(customButton('Done'))
		expect(saves().map(t => t.repeat)).toEqual([{freq: 'weekly'}])
		expect(sheet()).toBeNull()
	})
})
