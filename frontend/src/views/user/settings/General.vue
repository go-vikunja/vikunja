<template>
	<Card
		:title="$t('user.settings.sections.personalInformation')"
		class="general-settings"
		:loading="loading"
	>
		<div class="field-group">
			<FormField
				:label="$t('user.settings.general.name')"
				layout="two-col"
			>
				<FormInput
					v-model="settings.name"
					:disabled="isExternalUser"
					:placeholder="$t('user.settings.general.newName')"
					type="text"
					@keyup.enter="updateSettings"
				/>
			</FormField>
			<p
				v-if="isExternalUser"
				class="help"
			>
				{{ $t('user.settings.general.externalUserNameChange', {provider: authStore.info.auth_provider}) }}
			</p>
			<FormField
				:label="$t('user.settings.general.defaultProject')"
				layout="two-col"
			>
				<ProjectSearch v-model="defaultProject" />
			</FormField>
		</div>
	</Card>

	<Card
		:title="$t('user.settings.sections.taskAndNotifications')"
		class="general-settings section-block"
		:loading="loading"
	>
		<div class="field-group">
			<FormField
				:label="$t('user.settings.general.defaultView')"
				layout="two-col"
			>
				<FormSelect
					v-model="settings.frontend_settings.default_view"
					:options="defaultViewOptions"
				/>
			</FormField>
			<FormField
				:label="$t('user.settings.general.minimumPriority')"
				layout="two-col"
			>
				<FormSelect
					v-model="settings.frontend_settings.minimum_priority"
					:options="minimumPriorityOptions"
				/>
			</FormField>
			<FormField
				:label="$t('user.settings.general.defaultDueTime')"
				layout="two-col"
			>
				<FormInput
					v-model="settings.frontend_settings.default_due_time"
					type="time"
				/>
			</FormField>
			<p class="help">
				{{ $t('user.settings.general.defaultDueTimeDescription') }}
			</p>
			<FormField
				v-if="hasFilters"
				:label="$t('user.settings.general.filterUsedOnOverview')"
				layout="two-col"
			>
				<ProjectSearch
					v-model="filterUsedInOverview"
					:saved-filters-only="true"
				/>
			</FormField>
			<FormCheckbox
				v-model="settings.frontend_settings.show_last_viewed"
				:label="$t('user.settings.general.showLastViewed')"
			/>
			<FormCheckbox
				v-model="settings.email_reminders_enabled"
				:label="$t('user.settings.general.emailReminders')"
			/>
			<FormCheckbox
				v-model="settings.overdue_tasks_reminders_enabled"
				:label="$t('user.settings.general.overdueReminders')"
			/>
			<FormField
				v-if="settings.overdue_tasks_reminders_enabled"
				:label="$t('user.settings.general.overdueTasksRemindersTime')"
				layout="two-col"
			>
				<FormInput
					v-model="settings.overdue_tasks_reminders_time"
					type="time"
					@keyup.enter="updateSettings"
				/>
			</FormField>
		</div>
	</Card>

	<Card
		:title="$t('user.settings.sections.localization')"
		class="general-settings section-block"
		:loading="loading"
	>
		<div class="field-group">
			<FormField
				:label="$t('user.settings.general.language')"
				layout="two-col"
			>
				<FormSelect
					v-model="settings.language"
					:options="languageOptions"
				/>
			</FormField>
			<FormField
				:label="$t('user.settings.general.timezone')"
				layout="two-col"
			>
				<Multiselect
					v-model="timezoneObject"
					:placeholder="$t('user.settings.general.timezone')"
					:search-results="timezoneSearchResults"
					:show-empty="true"
					class="timezone-select"
					label="label"
					select-placeholder=""
					@search="searchTimezones"
				/>
			</FormField>
			<FormField
				:label="$t('user.settings.general.weekStart')"
				layout="two-col"
			>
				<FormSelect
					v-model.number="settings.week_start"
					v-cy="'weekStartSelect'"
					:options="weekStartOptions"
				/>
			</FormField>
			<FormField
				:label="$t('user.settings.general.dateDisplay')"
				layout="two-col"
			>
				<FormSelect
					v-model="settings.frontend_settings.date_display"
					:options="dateDisplayOptions"
				/>
			</FormField>
			<FormField
				v-if="settings.frontend_settings.date_display !== 'relative'"
				:label="$t('user.settings.general.timeFormat')"
				layout="two-col"
			>
				<FormSelect
					v-model="settings.frontend_settings.time_format"
					:options="timeFormatOptions"
				/>
			</FormField>
			<FormField
				v-if="timeTrackingEnabled"
				:label="$t('user.settings.general.timeTrackingDefaultStart')"
				layout="two-col"
			>
				<FormInput
					v-model="settings.frontend_settings.time_tracking_default_start"
					type="time"
				/>
			</FormField>
		</div>
	</Card>

	<Card
		:title="$t('user.settings.sections.appearance')"
		class="general-settings section-block"
		:loading="loading"
	>
		<div class="field-group">
			<FormField
				:label="$t('user.settings.appearance.title')"
				layout="two-col"
			>
				<FormSelect
					v-model="settings.frontend_settings.color_schema"
					:options="colorSchemeOptions"
				/>
			</FormField>
			<FormField
				:label="$t('user.settings.quickAddMagic.title')"
				layout="two-col"
			>
				<FormSelect
					v-model="settings.frontend_settings.quick_add_magic_mode"
					:options="quickAddMagicModeOptions"
				/>
			</FormField>
			<div
				v-if="settings.frontend_settings.quick_add_magic_mode !== PrefixMode.Disabled"
				class="field"
			>
				<label class="label">{{ $t('user.settings.general.quickAddDefaultReminders') }}</label>
				<p class="help">
					{{ $t('user.settings.general.quickAddDefaultRemindersDescription') }}
				</p>
				<p class="help">
					{{ $t('user.settings.general.quickAddDefaultRemindersHint') }}
				</p>
				<Reminders
					v-model="quick_add_default_reminders"
					:default-relative-to="REMINDER_PERIOD_RELATIVE_TO_TYPES.DUEDATE"
					:allow-absolute="false"
				/>
			</div>
			<FormField
				:label="$t('user.settings.general.defaultTaskRelationType')"
				layout="two-col"
			>
				<FormSelect
					v-model="settings.frontend_settings.default_task_relation_type"
					:options="defaultTaskRelationTypeOptions"
				/>
			</FormField>
			<FormCheckbox
				v-model="settings.frontend_settings.play_sound_when_done"
				:label="$t('user.settings.general.playSoundWhenDone')"
			/>
			<FormCheckbox
				v-model="settings.frontend_settings.allow_icon_changes"
				:label="$t('user.settings.general.allowIconChanges')"
			/>
			<FormCheckbox
				v-model="settings.frontend_settings.always_show_bucket_task_count"
				:label="$t('user.settings.general.alwaysShowBucketTaskCount')"
			/>
			<FormField
				:label="$t('user.settings.backgroundBrightness.title')"
				layout="two-col"
			>
				<FormInput
					v-model.number="settings.frontend_settings.background_brightness"
					type="number"
					min="0"
					max="100"
					@blur="enforceBackgroundBrightnessBounds"
				/>
			</FormField>
		</div>
	</Card>

	<Card
		v-if="isDesktop"
		:title="$t('user.settings.sections.desktop')"
		class="general-settings section-block"
		:loading="loading"
	>
		<div class="field-group">
			<FormField
				:label="$t('user.settings.desktop.quickEntryShortcut')"
				layout="two-col"
			>
				<ShortcutRecorder
					v-model="settings.frontend_settings.desktop_quick_entry_shortcut"
					@update:modelValue="updateSettings"
				/>
			</FormField>
		</div>
	</Card>

	<Card
		:title="$t('user.settings.sections.privacy')"
		class="general-settings section-block"
		:loading="loading"
	>
		<div class="field-group">
			<FormCheckbox
				v-model="settings.discoverable_by_name"
				:label="$t('user.settings.general.discoverableByName')"
			/>
			<FormCheckbox
				v-model="settings.discoverable_by_email"
				:label="$t('user.settings.general.discoverableByEmail')"
			/>
		</div>
	</Card>

	<div class="sticky-save">
		<CustomTransition name="fade">
			<XButton
				v-if="isDirty"
				v-cy="'saveGeneralSettings'"
				:loading="loading"
				class="is-fullwidth"
				@click="updateSettings()"
			>
				{{ $t('misc.save') }}
			</XButton>
		</CustomTransition>
	</div>
</template>


<script setup lang="ts">
import {useUpdateSettingsMutation} from '@/client/queries/account'
import {computed, watch, ref, onBeforeMount} from 'vue'
import {useI18n} from 'vue-i18n'
import isEqual from 'fast-deep-equal'

import {PrefixMode} from '@/modules/quickAddMagic'

import ProjectSearch from '@/components/tasks/partials/ProjectSearch.vue'
import Multiselect from '@/components/input/Multiselect.vue'
import CustomTransition from '@/components/misc/CustomTransition.vue'
import FormField from '@/components/input/FormField.vue'
import FormInput from '@/components/input/FormInput.vue'
import FormSelect from '@/components/input/FormSelect.vue'
import FormCheckbox from '@/components/input/FormCheckbox.vue'

import {SUPPORTED_LOCALES} from '@/i18n'
import {useQuery} from '@tanstack/vue-query'
import {timezonesQuery} from '@/client/queries/account'
import {formatDisplayDateFormat} from '@/helpers/time/formatDate'

import {useTitle} from '@/composables/useTitle'

import {useProjects} from '@/composables/useProjects'
import {useAuthStore} from '@/stores/auth'
import {useConfigStore} from '@/stores/config'
import {createUserSettingsDraft, taskRemindersFromSettings, type UserSettingsResponse} from '@/helpers/userSettings'
import {isSavedFilterProject} from '@/client/queries/projects'
import {DEFAULT_PROJECT_VIEW_SETTINGS} from '@/constants/projectView'
import {PRIORITIES} from '@/constants/priorities'
import {DATE_DISPLAY} from '@/constants/dateDisplay'
import {TIME_FORMAT} from '@/constants/timeFormat'
import {PRO_FEATURE} from '@/constants/proFeatures'
import {RELATION_KINDS} from '@/types/IRelationKind'
import {isDesktopApp} from '@/helpers/desktopAuth'
import ShortcutRecorder from '@/components/misc/ShortcutRecorder.vue'
import Reminders from '@/components/tasks/partials/Reminders.vue'
import {REMINDER_PERIOD_RELATIVE_TO_TYPES} from '@/types/IReminderPeriodRelativeTo'

defineOptions({name: 'UserSettingsGeneral'})

const isDesktop = isDesktopApp()

const {t} = useI18n({useScope: 'global'})
useTitle(() => `${t('user.settings.general.title')} - ${t('user.settings.title')}`)

const DEFAULT_PROJECT_ID = 0

const defaultViewOptions = computed(() =>
	Object.values(DEFAULT_PROJECT_VIEW_SETTINGS).map(view => ({
		value: view,
		label: t(`project.${view}.title`),
	})),
)

const minimumPriorityOptions = computed(() => [
	{value: PRIORITIES.LOW, label: t('task.priority.low')},
	{value: PRIORITIES.MEDIUM, label: t('task.priority.medium')},
	{value: PRIORITIES.HIGH, label: t('task.priority.high')},
	{value: PRIORITIES.URGENT, label: t('task.priority.urgent')},
	{value: PRIORITIES.DO_NOW, label: t('task.priority.doNow')},
])

const weekStartOptions = computed(() => [
	{value: 0, label: t('user.settings.general.weekStartSunday')},
	{value: 1, label: t('user.settings.general.weekStartMonday')},
	{value: 2, label: t('user.settings.general.weekStartTuesday')},
	{value: 3, label: t('user.settings.general.weekStartWednesday')},
	{value: 4, label: t('user.settings.general.weekStartThursday')},
	{value: 5, label: t('user.settings.general.weekStartFriday')},
	{value: 6, label: t('user.settings.general.weekStartSaturday')},
])

const dateDisplayOptions = computed(() => [
	{value: DATE_DISPLAY.RELATIVE, label: t('user.settings.general.dateDisplayOptions.relative')},
	{value: DATE_DISPLAY.MM_DD_YYYY, label: t('user.settings.general.dateDisplayOptions.mm-dd-yyyy')},
	{value: DATE_DISPLAY.DD_MM_YYYY, label: t('user.settings.general.dateDisplayOptions.dd-mm-yyyy')},
	{value: DATE_DISPLAY.YYYY_MM_DD, label: t('user.settings.general.dateDisplayOptions.yyyy-mm-dd')},
	{value: DATE_DISPLAY.MM_SLASH_DD_YYYY, label: t('user.settings.general.dateDisplayOptions.mm/dd/yyyy')},
	{value: DATE_DISPLAY.DD_SLASH_MM_YYYY, label: t('user.settings.general.dateDisplayOptions.dd/mm/yyyy')},
	{value: DATE_DISPLAY.YYYY_SLASH_MM_DD, label: t('user.settings.general.dateDisplayOptions.yyyy/mm/dd')},
	{value: DATE_DISPLAY.DAY_MONTH_YEAR, label: formatDisplayDateFormat(new Date(), DATE_DISPLAY.DAY_MONTH_YEAR, settings.value?.frontend_settings?.time_format)},
	{value: DATE_DISPLAY.WEEKDAY_DAY_MONTH_YEAR, label: formatDisplayDateFormat(new Date(), DATE_DISPLAY.WEEKDAY_DAY_MONTH_YEAR, settings.value?.frontend_settings?.time_format)},
])

const timeFormatOptions = computed(() => [
	{value: TIME_FORMAT.HOURS_12, label: t('user.settings.general.timeFormatOptions.12h')},
	{value: TIME_FORMAT.HOURS_24, label: t('user.settings.general.timeFormatOptions.24h')},
])

const colorSchemeOptions = computed(() => [
	{value: 'light', label: t('user.settings.appearance.colorScheme.light')},
	{value: 'auto', label: t('user.settings.appearance.colorScheme.system')},
	{value: 'dark', label: t('user.settings.appearance.colorScheme.dark')},
])

const quickAddMagicModeOptions = computed(() =>
	(Object.values(PrefixMode) as PrefixMode[]).map(mode => ({
		value: mode,
		label: t(`user.settings.quickAddMagic.${mode}`),
	})),
)

const defaultTaskRelationTypeOptions = computed(() =>
	RELATION_KINDS.map(kind => ({
		value: kind,
		label: t(`task.relation.kinds.${kind}`, 1),
	})),
)

const languageOptions = computed(() =>
	Object.entries(SUPPORTED_LOCALES)
		.map(([code, title]) => ({value: code, label: title}))
		.sort((a, b) => a.label.localeCompare(b.label)),
)

const authStore = useAuthStore()
const updateUserSettings = useUpdateSettingsMutation()
const configStore = useConfigStore()
const timeTrackingEnabled = computed(() => configStore.isProFeatureEnabled(PRO_FEATURE.TIME_TRACKING))

const settings = ref(createUserSettingsDraft(authStore.settings))

const quick_add_default_reminders = computed({
	get: () => taskRemindersFromSettings(settings.value.frontend_settings.quick_add_default_reminders),
	set: reminders => {
		settings.value.frontend_settings.quick_add_default_reminders = reminders.map(reminder => ({
			relative_period: reminder.relative_period,
		}))
	},
})

const initialSettings = ref<UserSettingsResponse>()
const isDirty = ref(false)

onBeforeMount(() => {
	initialSettings.value = JSON.parse(JSON.stringify(settings.value))
	isDirty.value = false
})

watch(
	() => settings.value,
	() => {
		isDirty.value = !isEqual(settings.value, initialSettings.value)
	},
	{deep: true},
)

function enforceBackgroundBrightnessBounds() {
	const value = Number(settings.value.frontend_settings.background_brightness)
    
	if (!value || isNaN(value)) {
		settings.value.frontend_settings.background_brightness = null
	} else if (value < 0) {
		settings.value.frontend_settings.background_brightness = 0
	} else if (value > 100) {
		settings.value.frontend_settings.background_brightness = 100
	}
}

function useAvailableTimezones(settingsRef: Ref<UserSettingsResponse>) {
	const zones = useQuery(timezonesQuery())
	const searchText = ref('')
	const availableTimezones = computed(() => [...(zones.data.value ?? [])].sort().map(value => ({value, label: value.replace(/_/g, ' ')})))
	const searchResults = computed(() => availableTimezones.value.filter(zone => zone.label.toLowerCase().includes(searchText.value.toLowerCase())))
	function search(query: string) {searchText.value = query}

	const timezoneObject = computed({
		get: () => ({
			value: settingsRef.value.timezone,
			label: settingsRef.value.timezone?.replace(/_/g, ' '),
		}),
		set: (obj) => {
			if (obj === null) {
				settingsRef.value.timezone = ''
				return
			}
			if (typeof obj === 'object' && 'value' in obj) {
				settingsRef.value.timezone = obj.value
			}
		},
	})

	return {
		availableTimezones,
		searchResults,
		search,
		timezoneObject,
	}
}

// Use the timezone composable and destructure its return values
const { 
	searchResults: timezoneSearchResults,
	search: searchTimezones, 
	timezoneObject,
} = useAvailableTimezones(settings)

const isExternalUser = computed(() => authStore.info?.is_local_user === false)

const projectList = useProjects()
const defaultProject = computed({
	get: () => projectList.projects[settings.value.default_project_id],
	set(l) {
		settings.value.default_project_id = l ? l.id : DEFAULT_PROJECT_ID
	},
})
const filterUsedInOverview = computed({
	get: () => projectList.projects[settings.value.frontend_settings.filter_id_used_on_overview],
	set(l) {
		settings.value.frontend_settings.filter_id_used_on_overview = l ? l.id : null
	},
})
const hasFilters = computed(() => projectList.projectsArray.some(isSavedFilterProject))
const loading = updateUserSettings.isPending

async function updateSettings() {
	try {
		await updateUserSettings.mutateAsync({
			settings: {...settings.value, ...(configStore.demo_mode_enabled ? {language: undefined} : {})},
		})
	} catch { return }
	initialSettings.value = JSON.parse(JSON.stringify(settings.value))
	isDirty.value = false
}
</script>

<style lang="scss" scoped>
.timezone-select {
	min-inline-size: 200px;
	flex-grow: 1;

	@media screen and (max-width: $tablet) {
		min-inline-size: unset;
	}
}

.section-block + .section-block {
	margin-block-start: 1.5rem;
}

.field-group {
	display: grid;
	grid-template-columns: 1fr;
}

.sticky-save {
	position: sticky;
	inset-block-end: 0;
	padding: .25rem 1rem 1rem;
}
</style>
