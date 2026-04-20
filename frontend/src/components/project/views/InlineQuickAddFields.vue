<template>
	<div
		v-if="hasAnyField"
		class="inline-quick-add-chip-bar"
	>
		<button
			v-for="chip in inlineChips"
			:key="chip.field"
			type="button"
			class="inline-quick-add-chip"
			:class="[`inline-quick-add-chip--${chip.modifier}`, {'is-set': chip.isSet}]"
			:disabled="disabled || undefined"
			@click.stop="toggleInlinePopup(chip.popup, $event)"
		>
			<span
				v-if="chip.colorValue"
				class="inline-quick-add-chip__swatch"
				:style="{background: chip.colorValue}"
			/>
			<Icon
				v-else
				:icon="chip.icon"
				class="inline-quick-add-chip__icon"
				:class="`inline-quick-add-chip__icon--${chip.modifier}`"
			/>
			<span>{{ chip.label }}</span>
			<span
				v-if="chip.isSet"
				class="inline-quick-add-chip__clear"
				@click.stop="clearField(chip.field)"
			>
				<Icon icon="times" />
			</span>
		</button>
	</div>

	<Teleport to="body">
		<div
			v-if="openPopup !== null"
			ref="popupRef"
			class="inline-quick-add-popup"
			:class="[
				`inline-quick-add-popup--${popupVariant}`,
				openPopup === 'reminder' ? 'inline-quick-add-popup--wide' : null,
				isPopupReady ? null : 'inline-quick-add-popup--measuring',
			]"
			:style="{top: `${popupPosition.top}px`, left: `${popupPosition.left}px`}"
		>
			<DatepickerInline
				v-if="openPopup === 'due'"
				v-model="fields.dueDate"
			/>
			<DatepickerInline
				v-else-if="openPopup === 'start'"
				v-model="fields.startDate"
			/>
			<ul
				v-else-if="openPopup === 'priority'"
				class="inline-quick-add-priority-options"
			>
				<li
					v-for="option in PRIORITY_OPTIONS"
					:key="option.value"
				>
					<button
						type="button"
						class="inline-quick-add-priority-option"
						:class="{'is-active': fields.priority === option.value}"
						@click="selectPriority(option.value)"
					>
						{{ $t(option.labelKey) }}
					</button>
				</li>
			</ul>
			<EditAssignees
				v-else-if="openPopup === 'assignee'"
				v-model="fields.assignees"
				:task-id="0"
				:project-id="projectId"
			/>
			<EditLabels
				v-else-if="openPopup === 'labels'"
				v-model="fields.labels"
				:task-id="0"
				:creatable="false"
			/>
			<Reminders
				v-else-if="openPopup === 'reminder'"
				v-model="fields.reminders"
				:default-relative-to="reminderDefaultRelativeTo"
			/>
			<DatepickerInline
				v-else-if="openPopup === 'endDate'"
				v-model="fields.endDate"
			/>
			<ColorPicker
				v-else-if="openPopup === 'color'"
				v-model="fields.color"
			/>
			<div
				v-else-if="openPopup === 'percentDone'"
				class="inline-quick-add-percent-done"
			>
				<input
					v-model.number="fields.percentDone"
					type="range"
					min="0"
					max="100"
					step="10"
					class="inline-quick-add-percent-done__slider"
				>
				<span class="inline-quick-add-percent-done__label">{{ fields.percentDone }}%</span>
			</div>
			<XButton
				v-if="openPopup !== 'priority'"
				class="inline-quick-add-popup__confirm"
				:shadow="false"
				@click="openPopup = null"
			>
				{{ $t('misc.confirm') }}
			</XButton>
		</div>
	</Teleport>
</template>

<script setup lang="ts">
import {computed, nextTick, onBeforeUnmount, onMounted, ref, watch} from 'vue'
import {useI18n} from 'vue-i18n'

import DatepickerInline from '@/components/input/DatepickerInline.vue'
import EditAssignees from '@/components/tasks/partials/EditAssignees.vue'
import EditLabels from '@/components/tasks/partials/EditLabels.vue'
import Reminders from '@/components/tasks/partials/Reminders.vue'
import ColorPicker from '@/components/input/ColorPicker.vue'
import XButton from '@/components/input/Button.vue'

import {formatDateShort} from '@/helpers/time/formatDate'
import {closeWhenClickedOutside} from '@/helpers/closeWhenClickedOutside'
import {DEFAULT_INLINE_QUICK_ADD_FIELDS} from '@/modelTypes/IUserSettings'
import type {IUser} from '@/modelTypes/IUser'
import type {ILabel} from '@/modelTypes/ILabel'
import type {ITaskReminder} from '@/modelTypes/ITaskReminder'
import type {IReminderPeriodRelativeTo} from '@/types/IReminderPeriodRelativeTo'
import {REMINDER_PERIOD_RELATIVE_TO_TYPES} from '@/types/IReminderPeriodRelativeTo'

import {useAuthStore} from '@/stores/auth'

defineProps<{
	projectId: number
	disabled: boolean
}>()

defineOptions({name: 'InlineQuickAddFields'})

const {t} = useI18n({useScope: 'global'})
const authStore = useAuthStore()

const PRIORITY_LABEL_KEYS: Record<number, string> = {
	1: 'low',
	2: 'medium',
	3: 'high',
	4: 'urgent',
	5: 'doNow',
}

// --- Field state ---

const fields = ref({
	dueDate: null as Date | null,
	startDate: null as Date | null,
	endDate: null as Date | null,
	priority: 0,
	assignees: [] as IUser[],
	labels: [] as ILabel[],
	reminders: [] as ITaskReminder[],
	color: '',
	percentDone: 0,
})

// --- Enabled fields ---

const enabledFields = computed(
	() => authStore.settings.frontendSettings.inlineQuickAddFields ?? DEFAULT_INLINE_QUICK_ADD_FIELDS,
)
const hasAnyField = computed(() => enabledFields.value.length > 0)

function isEnabled(field: string) {
	return enabledFields.value.includes(field as typeof enabledFields.value[number])
}

// --- Popup ---

type PopupKind = 'due' | 'start' | 'endDate' | 'assignee' | 'labels' | 'reminder' | 'priority' | 'color' | 'percentDone' | null
const openPopup = ref<PopupKind>(null)
const popupRef = ref<HTMLElement | null>(null)
const popupPosition = ref<{top: number, left: number}>({top: 0, left: 0})
const isPopupReady = ref(false)

const PRIORITY_OPTIONS = [
	{value: 0, labelKey: 'task.priority.unset'},
	{value: 1, labelKey: 'task.priority.low'},
	{value: 2, labelKey: 'task.priority.medium'},
	{value: 3, labelKey: 'task.priority.high'},
	{value: 4, labelKey: 'task.priority.urgent'},
	{value: 5, labelKey: 'task.priority.doNow'},
] as const

function selectPriority(value: number) {
	fields.value.priority = value
	openPopup.value = null
}

const popupVariant = computed(() => {
	if (openPopup.value === 'due' || openPopup.value === 'start' || openPopup.value === 'endDate') {
		return 'date'
	}
	return 'picker'
})

const reminderDefaultRelativeTo = computed<IReminderPeriodRelativeTo | null>(
	() => fields.value.dueDate !== null ? REMINDER_PERIOD_RELATIVE_TO_TYPES.DUEDATE : null,
)

// --- Popup positioning ---

const anchorChipRect = ref<DOMRect | null>(null)
let popupResizeObserver: ResizeObserver | null = null

function toggleInlinePopup(which: Exclude<PopupKind, null>, event: MouseEvent) {
	if (openPopup.value === which) {
		openPopup.value = null
		return
	}
	const chip = event.currentTarget as HTMLElement
	const rect = chip.getBoundingClientRect()
	anchorChipRect.value = rect
	popupPosition.value = {
		top: rect.bottom + 4,
		left: rect.left,
	}
	isPopupReady.value = false
	openPopup.value = which
	nextTick(() => {
		clampPopupToViewport()
		isPopupReady.value = true
		observePopupResize()
	})
}

function clampPopupToViewport() {
	const popup = popupRef.value
	const chipRect = anchorChipRect.value
	if (!popup || !chipRect) {
		return
	}
	const margin = 8
	const popupRect = popup.getBoundingClientRect()
	const top = chipRect.bottom + 4
	let left = chipRect.left

	if (left + popupRect.width + margin > window.innerWidth) {
		left = Math.max(margin, window.innerWidth - popupRect.width - margin)
	}

	popupPosition.value = {top, left}
}

function observePopupResize() {
	disconnectPopupResize()
	const popup = popupRef.value
	if (!popup || typeof ResizeObserver === 'undefined') {
		return
	}
	popupResizeObserver = new ResizeObserver(() => {
		requestAnimationFrame(() => clampPopupToViewport())
	})
	popupResizeObserver.observe(popup)
}

function disconnectPopupResize() {
	if (popupResizeObserver) {
		popupResizeObserver.disconnect()
		popupResizeObserver = null
	}
}

watch(openPopup, (value) => {
	if (value === null) {
		disconnectPopupResize()
		anchorChipRect.value = null
		isPopupReady.value = false
	}
})

watch(() => fields.value.assignees.length, (newLen, oldLen) => {
	if (newLen > oldLen && openPopup.value === 'assignee') {
		openPopup.value = null
	}
})

function onDocumentClick(e: MouseEvent) {
	if (openPopup.value !== null && popupRef.value) {
		closeWhenClickedOutside(e, popupRef.value, () => {
			openPopup.value = null
		})
	}
}

onMounted(() => document.addEventListener('click', onDocumentClick))
onBeforeUnmount(() => {
	document.removeEventListener('click', onDocumentClick)
	disconnectPopupResize()
})

// --- Chip rendering ---

type InlineChip = {
	field: string
	modifier: string
	icon: string
	popup: Exclude<PopupKind, null>
	isSet: boolean
	label: string
	colorValue?: string
}

const CHIP_CONFIG: Record<string, {modifier: string, icon: string, popup: Exclude<PopupKind, null>}> = {
	assignee: {modifier: 'assignee', icon: 'user', popup: 'assignee'},
	dueDate: {modifier: 'due', icon: 'calendar', popup: 'due'},
	startDate: {modifier: 'start', icon: 'play', popup: 'start'},
	endDate: {modifier: 'end', icon: 'stop', popup: 'endDate'},
	priority: {modifier: 'priority', icon: 'exclamation', popup: 'priority'},
	labels: {modifier: 'labels', icon: 'tags', popup: 'labels'},
	reminder: {modifier: 'reminder', icon: 'bell', popup: 'reminder'},
	color: {modifier: 'color', icon: 'fill-drip', popup: 'color'},
	percentDone: {modifier: 'percent', icon: 'percent', popup: 'percentDone'},
}

const assigneeChipLabel = computed(() => {
	const count = fields.value.assignees.length
	if (count === 0) return t('task.attributes.assignees')
	if (count === 1) return fields.value.assignees[0].name || fields.value.assignees[0].username
	return t('task.attributes.assigneesN', count)
})

const labelsChipLabel = computed(() => {
	const count = fields.value.labels.length
	if (count === 0) return t('task.attributes.labels')
	if (count === 1) return fields.value.labels[0].title
	return t('task.attributes.labelsN', count)
})

const reminderChipLabel = computed(() => {
	const count = fields.value.reminders.length
	if (count === 0) return t('task.attributes.reminders')
	return t('task.attributes.remindersN', count)
})

const inlineChips = computed<InlineChip[]>(() => {
	const chipLabel: Record<string, () => string> = {
		assignee: () => assigneeChipLabel.value,
		dueDate: () => fields.value.dueDate !== null ? formatDateShort(fields.value.dueDate) : t('task.attributes.dueDate'),
		startDate: () => fields.value.startDate !== null ? formatDateShort(fields.value.startDate) : t('task.attributes.startDate'),
		endDate: () => fields.value.endDate !== null ? formatDateShort(fields.value.endDate) : t('task.attributes.endDate'),
		priority: () => fields.value.priority !== 0 ? t(`task.priority.${PRIORITY_LABEL_KEYS[fields.value.priority]}`) : t('task.attributes.priority'),
		labels: () => labelsChipLabel.value,
		reminder: () => reminderChipLabel.value,
		color: () => t('task.attributes.color'),
		percentDone: () => fields.value.percentDone > 0 ? `${fields.value.percentDone}%` : t('task.attributes.percentDone'),
	}
	const chipIsSet: Record<string, () => boolean> = {
		assignee: () => fields.value.assignees.length > 0,
		dueDate: () => fields.value.dueDate !== null,
		startDate: () => fields.value.startDate !== null,
		endDate: () => fields.value.endDate !== null,
		priority: () => fields.value.priority !== 0,
		labels: () => fields.value.labels.length > 0,
		reminder: () => fields.value.reminders.length > 0,
		color: () => fields.value.color !== '',
		percentDone: () => fields.value.percentDone > 0,
	}

	return enabledFields.value.map(field => {
		const cfg = CHIP_CONFIG[field]
		return {
			field,
			modifier: cfg.modifier,
			icon: cfg.icon,
			popup: cfg.popup,
			isSet: chipIsSet[field](),
			label: chipLabel[field](),
			colorValue: field === 'color' && fields.value.color ? fields.value.color : undefined,
		}
	})
})

function clearField(field: string) {
	const clearMap: Record<string, () => void> = {
		assignee: () => { fields.value.assignees = [] },
		dueDate: () => { fields.value.dueDate = null },
		startDate: () => { fields.value.startDate = null },
		endDate: () => { fields.value.endDate = null },
		priority: () => { fields.value.priority = 0 },
		labels: () => { fields.value.labels = [] },
		reminder: () => { fields.value.reminders = [] },
		color: () => { fields.value.color = '' },
		percentDone: () => { fields.value.percentDone = 0 },
	}
	clearMap[field]?.()
}

// --- Public API ---

function getFieldValues() {
	return {
		dueDate: isEnabled('dueDate') && fields.value.dueDate !== null ? fields.value.dueDate : undefined,
		startDate: isEnabled('startDate') && fields.value.startDate !== null ? fields.value.startDate : undefined,
		endDate: isEnabled('endDate') && fields.value.endDate !== null ? fields.value.endDate : undefined,
		priority: isEnabled('priority') && fields.value.priority !== 0 ? fields.value.priority : undefined,
		hexColor: isEnabled('color') && fields.value.color !== '' ? fields.value.color : undefined,
		percentDone: isEnabled('percentDone') && fields.value.percentDone > 0 ? fields.value.percentDone : undefined,
		assignees: isEnabled('assignee') ? [...fields.value.assignees] : [],
		labels: isEnabled('labels') ? [...fields.value.labels] : [],
		reminders: isEnabled('reminder') ? [...fields.value.reminders] : [],
	}
}

function reset() {
	fields.value = {
		dueDate: null,
		startDate: null,
		endDate: null,
		priority: 0,
		assignees: [],
		labels: [],
		reminders: [],
		color: '',
		percentDone: 0,
	}
	openPopup.value = null
}

function containsElement(el: Element | null): boolean {
	return el !== null && popupRef.value?.contains(el) === true
}

const isPopupOpen = computed(() => openPopup.value !== null)
const taskColor = computed(() => fields.value.color)

defineExpose({
	getFieldValues,
	reset,
	containsElement,
	hasAnyField,
	isPopupOpen,
	taskColor,
})
</script>

<style lang="scss" scoped>
.inline-quick-add-chip-bar {
	display: grid;
	grid-template-columns: 1fr 1fr;
	gap: .375rem;
	margin-block-start: .5rem;
}

.inline-quick-add-chip {
	position: relative;
	display: inline-flex;
	align-items: center;
	gap: .4rem;
	padding: .3rem .65rem;
	border: 1px solid transparent;
	border-radius: $radius;
	background: transparent;
	color: var(--grey-700);
	font-size: .8rem;
	font-weight: 500;
	line-height: 1.2;
	cursor: pointer;
	transition: background-color $transition, color $transition, border-color $transition, box-shadow $transition;

	&:hover:not(:disabled) {
		background: var(--white);
		color: var(--grey-900);
		box-shadow: 0 1px 3px hsla(var(--grey-900-hsl), .12);
	}

	&:focus-visible {
		outline: none;
		box-shadow: 0 0 0 2px var(--primary-light);
	}

	&:disabled {
		cursor: not-allowed;
		opacity: .5;
	}
}

.inline-quick-add-chip__icon {
	font-size: .85rem;

	&--due {
		color: var(--danger);
	}

	&--start {
		color: var(--success);
	}

	&--priority {
		color: var(--warning);
	}

	&--assignee,
	&--labels,
	&--reminder {
		color: var(--primary);
	}

	&--end,
	&--color,
	&--percent {
		color: var(--grey-500);
	}
}

.inline-quick-add-chip__swatch {
	display: inline-block;
	inline-size: .85rem;
	block-size: .85rem;
	border-radius: .2rem;
	border: 1px solid var(--grey-300);
	flex-shrink: 0;
}

.inline-quick-add-chip__clear {
	display: none;
	align-items: center;
	justify-content: center;
	position: absolute;
	inset-inline-end: 0;
	inset-block-start: 0;
	block-size: 100%;
	padding-inline: .25rem;
	font-size: .65rem;
	background: transparent;
	border-radius: 0 $radius $radius 0;
	z-index: 1;
	color: inherit;

	&:hover {
		color: var(--danger);
	}
}

.inline-quick-add-chip:hover .inline-quick-add-chip__clear {
	display: flex;
}

.inline-quick-add-chip--due.is-set {
	background: var(--danger-light);
	color: var(--danger-dark);
}

.inline-quick-add-chip--start.is-set,
.inline-quick-add-chip--end.is-set {
	background: var(--success-light);
	color: var(--success-dark);
}

.inline-quick-add-chip--priority.is-set {
	background: var(--warning-light);
	color: var(--warning-dark);
}

.inline-quick-add-chip--assignee.is-set,
.inline-quick-add-chip--labels.is-set,
.inline-quick-add-chip--reminder.is-set,
.inline-quick-add-chip--percent.is-set,
.inline-quick-add-chip--color.is-set {
	background: var(--primary-light);
	color: var(--primary-dark);
}

.inline-quick-add-popup {
	position: fixed;
	z-index: 50;
	padding: .5rem;
	background: var(--white);
	border: 1px solid var(--grey-200);
	border-radius: $radius;
	box-shadow: var(--shadow-md);
}

.inline-quick-add-popup--picker.inline-quick-add-popup--wide {
	inline-size: min(28rem, calc(100vw - 2rem));
}

.inline-quick-add-popup--measuring {
	visibility: hidden;
}

.inline-quick-add-priority-options {
	display: flex;
	flex-direction: column;
	gap: .125rem;
	margin: 0;
	padding: 0;
	list-style: none;
}

.inline-quick-add-priority-option {
	inline-size: 100%;
	padding: .5rem .75rem;
	border: 0;
	border-radius: $radius;
	background: transparent;
	color: var(--text);
	text-align: start;
	font-size: .9rem;
	cursor: pointer;

	&:hover {
		background: var(--primary-light);
		color: var(--primary-dark);
	}

	&.is-active {
		background: var(--primary-light);
		color: var(--primary-dark);
		font-weight: 600;
	}
}

.inline-quick-add-percent-done {
	display: flex;
	align-items: center;
	gap: .75rem;
	padding: .5rem .25rem;

	&__slider {
		flex: 1;
		accent-color: var(--primary);
	}

	&__label {
		min-inline-size: 3rem;
		text-align: end;
		font-weight: 600;
		font-size: .9rem;
	}
}

.inline-quick-add-popup--picker {
	inline-size: min(18rem, calc(100vw - 2rem));

	:deep(.color-picker-container) {
		justify-content: start;
	}

	:deep(.popup) {
		inset-inline-start: 0 !important;
		inset-inline-end: auto !important;
		inline-size: 100%;
	}

	:deep(.reminder-options-popup) {
		inline-size: 100% !important;
		max-inline-size: 100%;
	}

	:deep(.reminder-options-popup .datepicker-inline) {
		flex-direction: row;
		gap: .75rem;
		align-items: stretch;
	}

	:deep(.reminder-options-popup .datepicker-inline__shortcuts) {
		display: flex;
		flex-direction: column;
		flex-shrink: 0;
	}

	:deep(.reminder-options-popup .datepicker-inline__shortcuts .datepicker__quick-select-date) {
		flex: 1 1 auto;
		block-size: auto;
	}

	:deep(.reminder-options-popup .flatpickr-container) {
		flex: 0 1 auto;
	}

	:deep(.reminder-options-popup .flatpickr-container > input) {
		display: none;
	}

	@media (width <= 520px) {
		:deep(.reminder-options-popup .datepicker-inline) {
			flex-direction: column;
		}
	}
}

.inline-quick-add-popup--date {
	display: flex;
	flex-direction: column;
	max-inline-size: calc(100vw - 1rem);
}

.inline-quick-add-popup--date :deep(.datepicker-inline) {
	flex-direction: row;
	gap: .75rem;
	align-items: stretch;
}

.inline-quick-add-popup--date :deep(.datepicker-inline__shortcuts) {
	display: flex;
	flex-direction: column;
	flex-shrink: 0;
}

.inline-quick-add-popup--date :deep(.datepicker-inline__shortcuts .datepicker__quick-select-date) {
	flex: 1 1 auto;
	block-size: auto;
}

.inline-quick-add-popup--date :deep(.flatpickr-container) {
	flex: 0 1 auto;
}

.inline-quick-add-popup--date :deep(.flatpickr-container > input) {
	display: none;
}

.inline-quick-add-popup__confirm {
	inline-size: 100%;
	margin-block-start: .5rem;
}

@media (width <= 520px) {
	.inline-quick-add-popup--date :deep(.datepicker-inline) {
		flex-direction: column;
	}
}
</style>
