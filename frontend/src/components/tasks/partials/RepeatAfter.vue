<template>
	<div class="repeat-after">
		<div class="repeat-field">
			<Popup
				:open="isOpen"
				placement="bottom-start"
				:anchor="triggerEl"
				sheet-on-mobile
				:sheet-title="t('task.repeat.sheetTitle')"
				@update:open="onOpenChange"
			>
				<template #trigger="{toggle}">
					<SimpleButton
						ref="trigger"
						class="repeat-field-trigger"
						aria-haspopup="dialog"
						:aria-expanded="isOpen ? 'true' : 'false'"
						:aria-controls="contentId"
						:aria-describedby="detailsId"
						:disabled="disabled || undefined"
						@click.stop="toggle()"
					>
						{{ summary }}
					</SimpleButton>
				</template>
				<template #content>
					<div
						:id="contentId"
						ref="content"
						class="repeat-popup"
						:role="asSheet ? undefined : 'dialog'"
						:aria-label="asSheet ? undefined : t('task.repeat.sheetTitle')"
						tabindex="-1"
					>
						<Card
							class="repeat-options-popup"
							:padding="false"
							:has-content="false"
						>
							<div
								v-if="repeatClass === 'locked'"
								class="repeat-locked-panel"
							>
								<p class="repeat-locked-summary">
									{{ summary }}
								</p>
								<p class="repeat-note">
									{{ t('task.repeat.lockedNote') }}
								</p>
							</div>
							<div
								v-else-if="mode !== 'custom'"
								class="repeat-options"
							>
								<BaseButton
									v-for="preset in presets"
									:key="preset.id"
									class="repeat-option"
									:class="{'currently-active': preset.id === currentPresetId}"
									:aria-current="preset.id === currentPresetId ? 'true' : undefined"
									:data-repeat-option="preset.id"
									@click="pickPreset(preset)"
								>
									{{ presetLabel(preset) }}
								</BaseButton>
								<BaseButton
									class="repeat-option"
									:class="{'currently-active': isCustomCurrent}"
									:aria-current="isCustomCurrent ? 'true' : undefined"
									:data-repeat-option="CUSTOM_OPTION"
									@click="openCustom"
								>
									{{ t('task.repeat.customOption') }}
								</BaseButton>
							</div>
							<div
								v-else-if="draft !== null"
								ref="customPanel"
								class="repeat-custom-panel"
								tabindex="-1"
							>
								<RepeatCustomForm
									v-model="draft"
									:due-date="dueDate"
									:anchor="anchor"
									:repeats-from-current-date="repeatsFromCurrentDate"
									:zone="zone"
									:preview-zone="previewZoneSetting"
									:week-start="weekStart"
									:stored-repeat="repeat"
									:disabled="disabled"
									@done="commitDraft"
									@back="backToPresets"
								/>
							</div>
							<p
								v-if="commitError"
								class="help is-danger repeat-commit-error"
								role="alert"
							>
								{{ t(commitError) }}
							</p>
						</Card>
					</div>
				</template>
			</Popup>
		</div>

		<div
			:id="detailsId"
			class="repeat-details"
		>
			<p
				v-if="repeatClass === 'foreign'"
				class="repeat-note"
			>
				{{ t('task.repeat.foreignNote') }}
			</p>
			<p
				v-else-if="repeatClass === 'locked'"
				class="repeat-note"
			>
				{{ t('task.repeat.lockedNote') }}
			</p>
			<p
				v-if="details.kind === 'dates'"
				class="repeat-preview"
			>
				<span>{{ t('task.repeat.custom.previewLabel') }}</span>{{ ' ' }}
				<template
					v-for="(date, index) in details.dates"
					:key="date"
				>
					<span class="repeat-preview-date">{{ index < details.dates.length - 1 ? `${date} ·` : date }}</span>{{ ' ' }}
				</template>
			</p>
			<p
				v-else-if="details.kind === 'due'"
				class="repeat-preview"
			>
				{{ t('task.repeat.currentDue', {date: details.date}) }}
			</p>
			<p
				v-else-if="details.kind === 'noZone'"
				class="repeat-note"
			>
				{{ t('task.repeat.custom.noPreview') }}
			</p>
			<p
				v-else-if="details.kind === 'noDueDate'"
				class="repeat-note"
			>
				{{ repeatsFromCurrentDate ? t('task.repeat.noDueDateCompletionNote') : t('task.repeat.noDueDateNote') }}
			</p>
			<p
				class="repeat-note"
				role="status"
			>
				{{ dueDateNote }}
			</p>
		</div>

		<FancyCheckbox
			class="repeat-from-completion"
			:model-value="repeatsFromCurrentDate"
			:disabled="disabled"
			@update:modelValue="setFromCompletion"
		>
			{{ t('task.repeat.fromCurrentDate') }}
		</FancyCheckbox>

		<Modal
			v-if="mode === 'confirm'"
			class="repeat-replace-dialog"
			:aria-labelledby="replaceTitleId"
			:aria-describedby="replaceTextId"
			@close="cancelReplace"
		>
			<div
				:id="replaceTitleId"
				class="modal-header repeat-replace-title"
			>
				{{ t('task.repeat.replaceTitle') }}
			</div>
			<div class="content">
				<p :id="replaceTextId">
					{{ t('task.repeat.replaceText', {summary}) }}
				</p>
			</div>
			<div class="actions repeat-replace-actions">
				<XButton
					variant="tertiary"
					class="has-text-danger repeat-replace-action"
					@click="cancelReplace"
				>
					{{ t('misc.cancel') }}
				</XButton>
				<XButton
					variant="primary"
					class="repeat-replace-action"
					:shadow="false"
					@click="confirmReplace"
				>
					{{ t('task.repeat.replace') }}
				</XButton>
			</div>
		</Modal>
	</div>
</template>

<script setup lang="ts">
import {computed, nextTick, onMounted, ref, useId, useTemplateRef, watch} from 'vue'
import {useI18n} from 'vue-i18n'
import dayjs from 'dayjs'
import utc from 'dayjs/plugin/utc'
import timezone from 'dayjs/plugin/timezone'

import BaseButton from '@/components/base/BaseButton.vue'
import XButton from '@/components/input/Button.vue'
import FancyCheckbox from '@/components/input/FancyCheckbox.vue'
import SimpleButton from '@/components/input/SimpleButton.vue'
import Card from '@/components/misc/Card.vue'
import Modal from '@/components/misc/Modal.vue'
import Popup from '@/components/misc/Popup.vue'
import RepeatCustomForm from '@/components/tasks/partials/RepeatCustomForm.vue'

import type {Task as ITask, TaskRepeat} from '@/client/generated'
import type {TaskInput} from '@/client/queries/taskMutations'
import {useIsMobile} from '@/composables/useIsMobile'
import {
	classifyRepeat,
	defaultCustomForm,
	dueDateOf,
	formToRepeat,
	normalizeRepeatForm,
	presetIdFor,
	previewZone,
	recurrenceZone,
	repeatPresets,
	repeatToForm,
	type RepeatForm,
	type RepeatPreset,
} from '@/helpers/rrule'
import {firstOccurrenceAfter, previewDates} from '@/helpers/rrulePreview'
import {describeRepeatRule, formatInZone, formatOrdinal, weekdayName, type DescribeOptions} from '@/helpers/rruleDescribe'
import {calculateNearestHours} from '@/helpers/time/calculateNearestHours'
import {useAuthStore} from '@/stores/auth'

const props = withDefaults(defineProps<{
	modelValue: ITask | undefined
	disabled?: boolean
	startOpen?: boolean
}>(), {
	disabled: false,
	startOpen: false,
})
const emit = defineEmits<{
	'update:modelValue': [value: TaskInput]
	// True while the presets, Custom form or replace confirmation is showing.
	'update:open': [open: boolean]
}>()
dayjs.extend(utc)
dayjs.extend(timezone)

type Mode = 'presets' | 'custom' | 'confirm'
type ReplaceIntent = {kind: 'preset', id: string, repeat: TaskRepeat | null} | {kind: 'custom'}
type Details =
	| {kind: 'none'}
	| {kind: 'dates', dates: string[]}
	| {kind: 'due', date: string}
	| {kind: 'noZone'}
	| {kind: 'noDueDate'}

const CUSTOM_OPTION = 'custom'
const FOCUSABLE = 'button:not([disabled]), input:not([disabled]), select:not([disabled]), textarea:not([disabled])'

const {t, locale} = useI18n({useScope: 'global'})
const authStore = useAuthStore()
const isMobile = useIsMobile()
// On phones Popup renders a sheet dialog, which is already named.
const asSheet = computed(() => isMobile.value)

const translate: DescribeOptions['t'] = (key, params, plural) => plural === undefined
	? t(key, params ?? {})
	: t(key, params ?? {}, plural)

const contentId = useId()
const detailsId = useId()
const replaceTitleId = useId()
const replaceTextId = useId()
const trigger = useTemplateRef<InstanceType<typeof SimpleButton>>('trigger')
const triggerEl = computed<HTMLElement | null>(() => trigger.value?.$el ?? null)
const content = useTemplateRef<HTMLElement>('content')
const customPanel = useTemplateRef<HTMLElement>('customPanel')

const isOpen = ref(false)
const mode = ref<Mode>('presets')
const draft = ref<RepeatForm | null>(null)
const pendingReplace = ref<ReplaceIntent | null>(null)
const commitError = ref<string | null>(null)
const dueDateSetTo = ref<Date | null>(null)
const openedAt = ref(new Date())
// A draft belongs to the task that was showing when it started; it is never written to another one.
const draftTaskId = ref<number | null>(null)

const taskId = computed(() => props.modelValue?.id ?? null)
const repeat = computed(() => props.modelValue?.repeat ?? null)
const dueDate = computed(() => dueDateOf(props.modelValue))
const repeatsFromCurrentDate = computed(() => props.modelValue?.repeats_from_current_date ?? false)

const zone = computed(() => recurrenceZone(authStore.settings))
const previewZoneSetting = computed(() => previewZone(authStore.settings))
const weekStart = computed(() => authStore.settings.week_start ?? 0)

// Without a due date: today in the recurrence zone, at the default time the date shortcuts use. The hour is
// picked from the zone's wall clock and applied in that zone, so it never moves the calendar day.
const anchor = computed(() => {
	if (dueDate.value !== null) {
		return dueDate.value
	}
	const zonedNow = dayjs(openedAt.value).tz(zone.value)
	const hour = calculateNearestHours(new Date(2000, 0, 1, zonedNow.hour(), zonedNow.minute()))
	return dayjs.tz(`${zonedNow.format('YYYY-MM-DD')} ${String(hour).padStart(2, '0')}:00`, zone.value).toDate()
})

const repeatClass = computed(() => classifyRepeat(repeat.value, dueDate.value))

const presets = computed(() => repeatPresets(anchor.value, zone.value))
const currentPresetId = computed(() => presetIdFor(repeat.value, anchor.value, zone.value))
const isCustomCurrent = computed(() => repeatClass.value === 'editable' && currentPresetId.value === null)

const describeOptions = computed<DescribeOptions>(() => ({
	t: translate,
	locale: locale.value,
	weekStart: weekStart.value,
	zone: zone.value,
	dueDate: dueDate.value,
}))

const summary = computed(() => repeat.value === null || repeatClass.value === 'none'
	? t('task.repeat.none')
	: describeRepeatRule(repeat.value, describeOptions.value))

const details = computed<Details>(() => {
	if (repeat.value === null || repeatClass.value === 'none') {
		return {kind: 'none'}
	}
	if (dueDate.value === null) {
		return {kind: 'noDueDate'}
	}
	const preview = previewDates(repeat.value, dueDate.value, new Date(), repeatsFromCurrentDate.value, previewZoneSetting.value)
	if (preview.dates !== null) {
		return {kind: 'dates', dates: preview.dates.map(date => formatInZone(date, zone.value, locale.value, 'date'))}
	}
	if (preview.hiddenReason === 'noZone') {
		return {kind: 'noZone'}
	}
	return {kind: 'due', date: formatInZone(dueDate.value, zone.value, locale.value, 'date')}
})

const dueDateNote = computed(() => {
	const set = dueDateSetTo.value
	if (set === null || dueDate.value?.getTime() !== set.getTime()) {
		return ''
	}
	return t('task.repeat.dueDateSet', {date: formatInZone(set, zone.value, locale.value, 'date')})
})

const editorOpen = computed(() => isOpen.value || mode.value === 'confirm')
watch(editorOpen, value => emit('update:open', value))

function presetLabel(preset: RepeatPreset): string {
	const params = preset.labelParams
	const weekday = typeof params.weekday === 'string' ? weekdayName(params.weekday, locale.value) : undefined
	switch (preset.id) {
		case 'weeklyOnDay':
			return t(preset.labelKey, {weekday})
		case 'monthlyOnDate':
			return t(preset.labelKey, {day: formatOrdinal(params.day as number, {t: translate, locale: locale.value})})
		case 'monthlyOnNthWeekday':
			return t(preset.labelKey, {ordinal: formatOrdinal(params.ordinal as number, {t: translate, locale: locale.value}), weekday})
		case 'yearlyOnDate': {
			// Noon UTC of a leap year names the stored month and day in any locale, including February 29.
			const date = new Date(Date.UTC(2024, (params.month as number) - 1, params.day as number, 12))
			return t(preset.labelKey, {date: formatInZone(date, 'UTC', locale.value, 'monthDay')})
		}
		default:
			return t(preset.labelKey)
	}
}

function focusTrigger() {
	trigger.value?.focus()
}

function focusInContent(selector: string | null = null) {
	const root = content.value
	if (!root) {
		return
	}
	const target = (selector === null ? null : root.querySelector<HTMLElement>(selector))
		?? root.querySelector<HTMLElement>(FOCUSABLE)
		?? root
	target.focus()
}

function optionSelector(key: string) {
	return `[data-repeat-option="${key}"]`
}

// Popup moves no focus itself. On phones the sheet's dialog takes focus and gives it back.
async function showPopup(focusSelector: string | null = null) {
	if (!isOpen.value) {
		isOpen.value = true
		markOpened()
	}
	await nextTick()
	if (isOpen.value) {
		focusInContent(focusSelector)
	}
}

function markOpened() {
	openedAt.value = new Date()
	commitError.value = null
	draftTaskId.value = taskId.value
}

function discardDraft() {
	draft.value = null
	pendingReplace.value = null
	mode.value = 'presets'
	commitError.value = null
}

function focusIsInsideOrLost() {
	const active = document.activeElement
	return !active || active === document.body || !active.isConnected || Boolean(content.value?.contains(active))
}

async function afterClose() {
	discardDraft()
	await nextTick()
	if (focusIsInsideOrLost()) {
		focusTrigger()
	}
}

function closePopup() {
	isOpen.value = false
	return afterClose()
}

function onOpenChange(value: boolean) {
	if (value) {
		isOpen.value = true
		markOpened()
		const selector = repeatClass.value === 'locked' ? null : optionSelector(presets.value[0]?.id ?? CUSTOM_OPTION)
		showPopup(selector)
		return
	}
	isOpen.value = false
	afterClose()
}

function open() {
	focusTrigger()
	showPopup(optionSelector(presets.value[0]?.id ?? CUSTOM_OPTION))
}

function isCurrentDraftTask(task: ITask | undefined): task is ITask {
	return task !== undefined && task.id === draftTaskId.value
}

function commit(newRepeat: TaskRepeat | null): boolean {
	const task = props.modelValue
	if (!isCurrentDraftTask(task)) {
		return false
	}
	let dueDateValue = task.due_date
	if (newRepeat !== null && dueDate.value === null) {
		const first = firstOccurrenceAfter(newRepeat, anchor.value, new Date(), zone.value)
		if (first === null) {
			commitError.value = 'task.repeat.error.noOccurrence'
			return false
		}
		dueDateSetTo.value = first
		dueDateValue = first.toISOString()
	}
	emit('update:modelValue', {
		...task,
		repeat: newRepeat,
		repeats_from_current_date: task.repeats_from_current_date ?? false,
		due_date: dueDateValue,
	})
	return true
}

function requestReplace(intent: ReplaceIntent) {
	pendingReplace.value = intent
	mode.value = 'confirm'
	// The confirmation is a modal dialog: on wide screens it replaces the popover, on phones it stacks on the sheet.
	if (!asSheet.value) {
		isOpen.value = false
	}
}

function pickPreset(preset: RepeatPreset) {
	if (repeatClass.value === 'locked') {
		return
	}
	if (repeatClass.value === 'foreign') {
		requestReplace({kind: 'preset', id: preset.id, repeat: preset.repeat})
		return
	}
	// Re-picking the current preset saves nothing, unless the repeat still needs its due date.
	const alreadySet = preset.id === currentPresetId.value && (dueDate.value !== null || preset.repeat === null)
	if (alreadySet || commit(preset.repeat)) {
		closePopup()
	}
}

function openCustom() {
	if (repeatClass.value === 'locked') {
		return
	}
	if (repeatClass.value === 'foreign') {
		requestReplace({kind: 'custom'})
		return
	}
	startCustom(repeatClass.value === 'editable' && repeat.value !== null
		? repeatToForm(repeat.value, dueDate.value, zone.value)
		: defaultCustomForm(anchor.value, zone.value))
}

async function startCustom(form: RepeatForm) {
	draft.value = form
	mode.value = 'custom'
	commitError.value = null
	// On phones the form itself takes focus: even a moment on its first field opens the soft keyboard.
	await showPopup(asSheet.value ? '.repeat-custom-panel' : null)
	if (!asSheet.value) {
		customPanel.value?.querySelector<HTMLElement>(FOCUSABLE)?.focus()
	}
}

async function backToPresets() {
	discardDraft()
	await nextTick()
	focusInContent(optionSelector(CUSTOM_OPTION))
}

function commitDraft() {
	if (draft.value === null) {
		return
	}
	if (commit(formToRepeat(normalizeRepeatForm(draft.value), dueDate.value, zone.value))) {
		closePopup()
	}
}

function cancelReplace() {
	const intent = pendingReplace.value
	pendingReplace.value = null
	mode.value = 'presets'
	const origin = intent?.kind === 'preset' ? intent.id : CUSTOM_OPTION
	showPopup(optionSelector(origin))
}

function confirmReplace() {
	const intent = pendingReplace.value
	pendingReplace.value = null
	if (intent === null) {
		return
	}
	if (intent.kind === 'custom') {
		startCustom(defaultCustomForm(anchor.value, zone.value))
		return
	}
	mode.value = 'presets'
	if (commit(intent.repeat)) {
		closePopup()
		return
	}
	showPopup()
}

// Sends the stored rule back whole: a server read-only rule only accepts an unchanged echo.
function setFromCompletion(value: boolean) {
	if (!props.modelValue) {
		return
	}
	emit('update:modelValue', {...props.modelValue, repeats_from_current_date: value})
}

watch(taskId, () => {
	isOpen.value = false
	discardDraft()
	dueDateSetTo.value = null
	draftTaskId.value = null
})

onMounted(() => {
	if (props.startOpen && !props.disabled) {
		open()
	}
})

defineExpose({
	focus: focusTrigger,
})
</script>

<style lang="scss" scoped>
.repeat-after {
	display: flex;
	flex-direction: column;
	align-items: flex-start;
	gap: .25rem;
	margin-block-start: .25rem;
	max-inline-size: 100%;
}

.repeat-field {
	max-inline-size: 100%;
}

.repeat-field-trigger {
	min-block-size: 2.5rem;
	min-inline-size: 2.5rem;
	max-inline-size: 100%;
	overflow-wrap: anywhere;

	&:focus-visible {
		outline: 2px solid var(--primary);
		outline-offset: 2px;
	}

	&:disabled {
		cursor: default;

		&:hover {
			background: transparent;
		}
	}
}

.repeat-popup:focus,
.repeat-custom-panel:focus {
	outline: none;
}

.repeat-options-popup {
	inline-size: min(22rem, calc(100vw - 2rem));
	z-index: 99;

	.bottom-sheet & {
		inline-size: 100%;
		border: 0;
		box-shadow: none;
		border-radius: 0;
		padding-block-end: .5rem;

		.repeat-option {
			padding: .75rem 1rem;
			font-size: 1rem;
		}
	}
}

.repeat-options {
	display: flex;
	flex-direction: column;
}

.repeat-option {
	min-block-size: 2.5rem;
	inline-size: 100%;
	padding: .5rem;
	font-size: .85rem;
	color: var(--text);
	text-align: start;
	transition: background-color $transition;

	&:hover {
		background: var(--grey-100);
	}

	&.currently-active {
		color: var(--primary);
	}
}

.repeat-locked-panel {
	padding: .75rem;

	p {
		margin: 0;
	}
}

.repeat-locked-summary {
	margin-block-end: .25rem;
}

.repeat-custom-panel {
	padding: .75rem;
}

.repeat-commit-error {
	margin: 0;
	padding: 0 .75rem .75rem;
}

// Outranks Bulma's .help.is-danger (--danger, 3.46:1 on white) regardless of stylesheet order.
.repeat-popup .repeat-commit-error {
	color: var(--danger-text);
}

.repeat-details {
	display: flex;
	flex-direction: column;
	gap: .125rem;
	padding-inline: .5rem;
	font-size: .85rem;
	// --text-light falls below 4.5:1 in the dark theme.
	color: var(--grey-600);

	p {
		margin: 0;
	}
}

// A date never breaks across lines, and a wrapped line never starts with a separator.
.repeat-preview {
	display: flex;
	flex-wrap: wrap;
	column-gap: .3em;
}

.repeat-preview-date {
	white-space: nowrap;
}

.repeat-from-completion {
	margin-inline-start: .5rem;
}

.repeat-replace-title {
	margin-block-end: .5rem;
	font-size: 1.5rem;
	font-weight: 700;
}

.repeat-replace-actions {
	display: flex;
	flex-wrap: wrap;
	justify-content: center;
	gap: .5rem;
}

.repeat-replace-action {
	min-block-size: 2.5rem;
	min-inline-size: 2.5rem;
}
</style>

<style lang="scss">
// Upstream hides #app beneath a phone modal so nothing shows through its scrim. The sheet is teleported outside
// #app, so it is hidden the same way while the replace confirmation is stacked on it; it stays open underneath.
body:has(dialog[open].repeat-replace-dialog) dialog[open].bottom-sheet:has(.repeat-popup) {
	opacity: 0;
}
</style>
