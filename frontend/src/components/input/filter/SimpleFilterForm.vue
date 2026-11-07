<script setup lang="ts">
import {computed, ref, useId, watch} from 'vue'
import {useI18n} from 'vue-i18n'
import {isDateMath, isSimpleUsername, type SimpleFilter} from '@/helpers/simpleFilter'
import type {Priority} from '@/constants/priorities'
import {formatDate} from '@/helpers/time/formatDate'
import {createDateFromString} from '@/helpers/time/createDateFromString'
import FancyCheckbox from '@/components/input/FancyCheckbox.vue'
import DatepickerWithRange from '@/components/date/DatepickerWithRange.vue'
import PrioritySelect from '@/components/tasks/partials/PrioritySelect.vue'
import EditLabels from '@/components/tasks/partials/EditLabels.vue'
import EditAssignees from '@/components/tasks/partials/EditAssignees.vue'
import ProjectSearch from '@/components/tasks/partials/ProjectSearch.vue'
import XButton from '@/components/input/Button.vue'
import BaseButton from '@/components/base/BaseButton.vue'
import Icon from '@/components/misc/Icon'
import {useLabels} from '@/composables/useLabels'
import {useProjects} from '@/composables/useProjects'
import {normalizeProject} from '@/client/queries/projects'
import type {User} from '@/client/generated'

const props = defineProps<{
	projectId?: number
}>()
const model = defineModel<SimpleFilter>({required: true})
const {t} = useI18n({useScope: 'global'})
const id = useId()

function update(patch: Partial<SimpleFilter>) {
	model.value = {
		...model.value,
		...patch,
	}
}

const dates = ref<{
	dateFrom: Date | string | null
	dateTo: Date | string | null
}>({
	dateFrom: null,
	dateTo: null,
})
// `before` keeps the picker's From locally, so unticking overdue restores the range.
watch(() => model.value.due, due => {
	const from = serializeDate(dates.value.dateFrom)
	const to = serializeDate(dates.value.dateTo)
	if (due.kind === 'between' ? from === due.from && to === due.to
		: due.kind === 'before' ? to === due.to || (due.to === 'now' && !to)
			: !from && !to) {
		return
	}
	dates.value = {
		dateFrom: due.kind === 'between' ? toPickerDate(due.from) : null,
		dateTo: due.kind === 'any' || due.to === 'now' ? null : toPickerDate(due.to),
	}
}, {immediate: true})

// The picker shows strings verbatim and Dates as UTC ISO; this is the format it writes itself.
function toPickerDate(value: string | undefined): string | null {
	if (!value) {
		return null
	}
	return isDateMath(value) ? value : formatDate(value, 'YYYY-MM-DD HH:mm')
}

function serializeDate(value: Date | string | null): string {
	if (!value) {
		return ''
	}
	if (typeof value === 'string' && isDateMath(value)) {
		return value
	}
	const date = createDateFromString(value)
	return Number.isNaN(date.getTime()) ? '' : date.toISOString()
}

function setDates(value: typeof dates.value) {
	dates.value = value
	const from = serializeDate(value.dateFrom)
	const to = serializeDate(value.dateTo)
	const includeOverdue = model.value.due.kind === 'before'
	if (!value.dateFrom && !value.dateTo) {
		update({due: includeOverdue ? {kind: 'before', to: 'now'} : {kind: 'any'}})
	} else if (includeOverdue && to) {
		update({due: {kind: 'before', to}})
	} else if (!includeOverdue && from && to) {
		update({due: {kind: 'between', from, to}})
	}
}

function setOverdue(includeOverdue: boolean) {
	const from = serializeDate(dates.value.dateFrom)
	const to = serializeDate(dates.value.dateTo)
	if (includeOverdue) {
		update({due: {kind: 'before', to: to || 'now'}})
	} else {
		update({due: to && to !== 'now' ? {kind: 'between', from: from || 'now', to} : {kind: 'any'}})
	}
}

// The overdue query ignores From, so don't show one.
const untilText = computed(() => {
	const {due} = model.value
	return due.kind === 'before' && due.to !== 'now'
		? t('filters.simple.until', {date: toPickerDate(due.to)})
		: null
})

const {getLabelById} = useLabels()
const labels = computed(() => model.value.labelIds.map(id => getLabelById(id) ?? {
	id,
	title: t('filters.simple.unknownLabel', {id}),
}))

// The query only stores usernames; negative ids keep them distinct without colliding with real users.
const assignees = computed(() => model.value.assignees.map((username, index) => ({
	id: -(index + 1),
	username,
})))

function isSelectableAssignee(user: User) {
	return user.username !== undefined && isSimpleUsername(user.username) && !model.value.assignees.includes(user.username)
}

const projectList = useProjects()
const projects = computed(() => model.value.projectIds.map(id => projectList.projects[id] ?? normalizeProject({
	id,
	title: t('filters.simple.unknownProject', {id}),
})))

const priority = computed({
	get: () => model.value.minPriority ?? 0,
	set: (value: number) => update({minPriority: value === 0 ? null : value as Priority}),
})
</script>

<template>
	<div class="simple-filter-form">
		<div class="field">
			<FancyCheckbox
				:model-value="model.showDone"
				@update:modelValue="update({showDone: $event})"
			>
				{{ t('filters.simple.showDone') }}
			</FancyCheckbox>
		</div>
		<div
			class="field"
			role="group"
			:aria-labelledby="`${id}-due`"
		>
			<div
				:id="`${id}-due`"
				class="label"
			>
				{{ t('task.attributes.dueDate') }}
			</div>
			<div class="due-date">
				<DatepickerWithRange
					:model-value="dates"
					@update:modelValue="setDates"
				>
					<template #trigger="{toggle, buttonText}">
						<XButton
							variant="secondary"
							@click.prevent="toggle()"
						>
							{{ untilText ?? buttonText }}
						</XButton>
					</template>
				</DatepickerWithRange>
				<BaseButton
					v-if="dates.dateFrom || dates.dateTo"
					v-tooltip="t('input.datepicker.clear')"
					class="clear-due-date"
					:aria-label="t('input.datepicker.clear')"
					@click="setDates({dateFrom: null, dateTo: null})"
				>
					<Icon icon="times" />
				</BaseButton>
			</div>
		</div>
		<div class="field">
			<FancyCheckbox
				:model-value="model.due.kind === 'before'"
				@update:modelValue="setOverdue"
			>
				{{ t('task.show.overdue') }}
			</FancyCheckbox>
		</div>
		<div
			class="field"
			role="group"
			:aria-labelledby="`${id}-labels`"
		>
			<div
				:id="`${id}-labels`"
				class="label"
			>
				{{ t('task.attributes.labels') }}
			</div>
			<EditLabels
				:model-value="labels"
				:creatable="false"
				:aria-label="t('task.attributes.labels')"
				@update:modelValue="update({labelIds: $event.map(label => label.id!)})"
			/>
		</div>
		<div
			class="field"
			role="group"
			:aria-labelledby="`${id}-assignees`"
		>
			<div
				:id="`${id}-assignees`"
				class="label"
			>
				{{ t('task.attributes.assignees') }}
			</div>
			<EditAssignees
				:model-value="assignees"
				:project-id="props.projectId"
				:filter="isSelectableAssignee"
				:aria-label="t('task.attributes.assignees')"
				@update:modelValue="update({assignees: ($event ?? []).map(user => user.username!)})"
			/>
		</div>
		<div
			v-if="!props.projectId"
			class="field"
			role="group"
			:aria-labelledby="`${id}-projects`"
		>
			<div
				:id="`${id}-projects`"
				class="label"
			>
				{{ t('project.projects') }}
			</div>
			<ProjectSearch
				multiple
				:model-value="projects"
				:aria-label="t('project.projects')"
				@update:modelValue="update({projectIds: $event.map(project => project.id)})"
			/>
		</div>
		<div
			class="field"
			role="group"
			:aria-labelledby="`${id}-priority`"
		>
			<div
				:id="`${id}-priority`"
				class="label"
			>
				{{ t('filters.simple.minPriority') }}
			</div>
			<PrioritySelect v-model="priority" />
		</div>
	</div>
</template>

<style lang="scss" scoped>
.due-date {
	display: flex;
	align-items: center;
}

.clear-due-date {
	color: var(--danger);
	padding-inline-start: .5rem;
	line-height: 1;
}
</style>
