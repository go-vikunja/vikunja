<template>
	<Card
		:title="$t('user.settings.sections.personalInformation')"
		class="general-settings"
		:loading="loading"
	>
		<div class="field-group">
			<div class="field">
				<label
					:for="`newName${id}`"
					class="two-col"
				>
					<span>
						{{ $t('user.settings.general.name') }}
					</span>
					<input
						:id="`newName${id}`"
						v-model="settings.name"
						:disabled="isExternalUser"
						class="input"
						:placeholder="$t('user.settings.general.newName')"
						type="text"
						@keyup.enter="updateSettings"
					>
				</label>
				<p
					v-if="isExternalUser"
					class="help"
				>
					{{ $t('user.settings.general.externalUserNameChange', {provider: authStore.info.authProvider}) }}
				</p>
			</div>
			<div class="field">
				<label class="two-col">
					<span>
						{{ $t('user.settings.general.defaultProject') }}
					</span>
					<ProjectSearch v-model="defaultProject" />
				</label>
			</div>
		</div>
	</Card>

	<Card
		:title="$t('user.settings.sections.taskAndNotifications')"
		class="general-settings section-block"
		:loading="loading"
	>
		<div class="field-group">
			<div class="field">
				<label class="two-col">
					<span>
						{{ $t('user.settings.general.defaultView') }}
					</span>
					<div class="select">
						<select v-model="settings.frontendSettings.defaultView">
							<option
								v-for="view in DEFAULT_PROJECT_VIEW_SETTINGS"
								:key="view"
								:value="view"
							>
								{{ $t(`project.${view}.title`) }}
							</option>
						</select>
					</div>
				</label>
			</div>
			<div class="field">
				<label class="two-col">
					<span>
						{{ $t('user.settings.general.minimumPriority') }}
					</span>
					<div class="select">
						<select v-model="settings.frontendSettings.minimumPriority">
							<option :value="PRIORITIES.LOW">
								{{ $t('task.priority.low') }}
							</option>
							<option :value="PRIORITIES.MEDIUM">
								{{ $t('task.priority.medium') }}
							</option>
							<option :value="PRIORITIES.HIGH">
								{{ $t('task.priority.high') }}
							</option>
							<option :value="PRIORITIES.URGENT">
								{{ $t('task.priority.urgent') }}
							</option>
							<option :value="PRIORITIES.DO_NOW">
								{{ $t('task.priority.doNow') }}
							</option>
						</select>
					</div>
				</label>
			</div>
			<div
				v-if="hasFilters"
				class="field"
			>
				<label class="two-col">
					<span>
						{{ $t('user.settings.general.filterUsedOnOverview') }}
					</span>
					<ProjectSearch
						v-model="filterUsedInOverview"
						:saved-filters-only="true"
					/>
				</label>
			</div>
			<div class="field">
				<label class="checkbox">
					<input
						v-model="settings.emailRemindersEnabled"
						type="checkbox"
					>
					{{ $t('user.settings.general.emailReminders') }}
				</label>
			</div>
			<div class="field">
				<label class="checkbox">
					<input
						v-model="settings.overdueTasksRemindersEnabled"
						type="checkbox"
					>
					{{ $t('user.settings.general.overdueReminders') }}
				</label>
			</div>
			<div
				v-if="settings.overdueTasksRemindersEnabled"
				class="field"
			>
				<label
					for="overdueTasksReminderTime"
					class="two-col"
				>
					<span>
						{{ $t('user.settings.general.overdueTasksRemindersTime') }}
					</span>
					<input
						id="overdueTasksReminderTime"
						v-model="settings.overdueTasksRemindersTime"
						class="input"
						type="time"
						@keyup.enter="updateSettings"
					>
				</label>
			</div>
		</div>
	</Card>

	<Card
		:title="$t('user.settings.sections.localization')"
		class="general-settings section-block"
		:loading="loading"
	>
		<div class="field-group">
			<div class="field">
				<label class="two-col">
					<span>
						{{ $t('user.settings.general.language') }}
					</span>
					<div class="select">
						<select v-model="settings.language">
							<option
								v-for="lang in availableLanguageOptions"
								:key="lang.code"
								:value="lang.code"
							>{{ lang.title }}
							</option>
						</select>
					</div>
				</label>
			</div>
			<div class="field">
				<label class="two-col">
					<span>
						{{ $t('user.settings.general.timezone') }}
					</span>
					<Multiselect
						v-model="timezoneObject"
						:placeholder="$t('user.settings.general.timezone')"
						:search-results="timezoneSearchResults"
						:show-empty="true"
						class="timezone-select"
						label="label"
						@search="searchTimezones"
					/>
				</label>
			</div>
			<div class="field">
				<label class="two-col">
					<span>
						{{ $t('user.settings.general.weekStart') }}
					</span>
					<div class="select">
						<select v-model.number="settings.weekStart">
							<option value="0">{{ $t('user.settings.general.weekStartSunday') }}</option>
							<option value="1">{{ $t('user.settings.general.weekStartMonday') }}</option>
						</select>
					</div>
				</label>
			</div>
			<div class="field">
				<label class="two-col">
					<span>
						{{ $t('user.settings.general.dateDisplay') }}
					</span>
					<div class="select">
						<select v-model="settings.frontendSettings.dateDisplay">
							<option
								v-for="(label, value) in dateDisplaySettings"
								:key="value"
								:value="value"
							>{{ label }}</option>
						</select>
					</div>
				</label>
			</div>
		</div>
	</Card>

	<Card
		:title="$t('user.settings.sections.appearance')"
		class="general-settings section-block"
		:loading="loading"
	>
		<div class="field-group">
			<div class="field">
				<label class="two-col">
					<span>
						{{ $t('user.settings.appearance.title') }}
					</span>
					<div class="select">
						<select v-model="settings.frontendSettings.colorSchema">
							<option
								v-for="(title, schemeId) in colorSchemeSettings"
								:key="schemeId"
								:value="schemeId"
							>
								{{ title }}
							</option>
						</select>
					</div>
				</label>
			</div>
			<div class="field">
				<label class="two-col">
					<span>
						{{ $t('user.settings.quickAddMagic.title') }}
					</span>
					<div class="select">
						<select v-model="settings.frontendSettings.quickAddMagicMode">
							<option
								v-for="set in PrefixMode"
								:key="set"
								:value="set"
							>
								{{ $t(`user.settings.quickAddMagic.${set}`) }}
							</option>
						</select>
					</div>
				</label>
			</div>
			<div class="field">
				<label class="checkbox">
					<input
						v-model="settings.frontendSettings.playSoundWhenDone"
						type="checkbox"
					>
					{{ $t('user.settings.general.playSoundWhenDone') }}
				</label>
			</div>
			<div class="field">
				<label class="checkbox">
					<input
						v-model="settings.frontendSettings.allowIconChanges"
						type="checkbox"
					>
					{{ $t('user.settings.general.allowIconChanges') }}
				</label>
			</div>
		</div>
	</Card>
	
	<Card
		:title="$t('user.settings.sections.privacy')"
		class="general-settings section-block"
		:loading="loading"
	>
		<div class="field-group">
			<div class="field">
				<label class="checkbox">
					<input
						v-model="settings.discoverableByName"
						type="checkbox"
					>
					{{ $t('user.settings.general.discoverableByName') }}
				</label>
			</div>
			<div class="field">
				<label class="checkbox">
					<input
						v-model="settings.discoverableByEmail"
						type="checkbox"
					>
					{{ $t('user.settings.general.discoverableByEmail') }}
				</label>
			</div>
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
import {computed, watch, ref, onBeforeMount} from 'vue'
import {useI18n} from 'vue-i18n'
import isEqual from 'fast-deep-equal'

import {PrefixMode} from '@/modules/parseTaskText'

import ProjectSearch from '@/components/tasks/partials/ProjectSearch.vue'
import Multiselect from '@/components/input/Multiselect.vue'
import CustomTransition from '@/components/misc/CustomTransition.vue'

import {SUPPORTED_LOCALES} from '@/i18n'
import {createRandomID} from '@/helpers/randomId'
import {AuthenticatedHTTPFactory} from '@/helpers/fetcher'
import {formatDisplayDateFormat} from '@/helpers/time/formatDate'

import {useTitle} from '@/composables/useTitle'

import {useProjectStore} from '@/stores/projects'
import {useAuthStore} from '@/stores/auth'
import type {IUserSettings} from '@/modelTypes/IUserSettings'
import {isSavedFilter} from '@/services/savedFilter'
import {DEFAULT_PROJECT_VIEW_SETTINGS} from '@/modelTypes/IProjectView'
import {PRIORITIES} from '@/constants/priorities'
import {DATE_DISPLAY} from '@/constants/dateDisplay'

defineOptions({name: 'UserSettingsGeneral'})

const {t} = useI18n({useScope: 'global'})
useTitle(() => `${t('user.settings.general.title')} - ${t('user.settings.title')}`)

const DEFAULT_PROJECT_ID = 0

const colorSchemeSettings = computed(() => ({
	light: t('user.settings.appearance.colorScheme.light'),
	auto: t('user.settings.appearance.colorScheme.system'),
	dark: t('user.settings.appearance.colorScheme.dark'),
}))

const dateDisplaySettings = computed(() => ({
	[DATE_DISPLAY.RELATIVE]: t('user.settings.general.dateDisplayOptions.relative'),
	[DATE_DISPLAY.MM_DD_YYYY]: t('user.settings.general.dateDisplayOptions.mm-dd-yyyy'),
	[DATE_DISPLAY.DD_MM_YYYY]: t('user.settings.general.dateDisplayOptions.dd-mm-yyyy'),
	[DATE_DISPLAY.YYYY_MM_DD]: t('user.settings.general.dateDisplayOptions.yyyy-mm-dd'),
	[DATE_DISPLAY.MM_SLASH_DD_YYYY]: t('user.settings.general.dateDisplayOptions.mm/dd/yyyy'),
	[DATE_DISPLAY.DD_SLASH_MM_YYYY]: t('user.settings.general.dateDisplayOptions.dd/mm/yyyy'),
	[DATE_DISPLAY.YYYY_SLASH_MM_DD]: t('user.settings.general.dateDisplayOptions.yyyy/mm/dd'),
	[DATE_DISPLAY.DAY_MONTH_YEAR]: formatDisplayDateFormat(new Date(), DATE_DISPLAY.DAY_MONTH_YEAR),
	[DATE_DISPLAY.WEEKDAY_DAY_MONTH_YEAR]: formatDisplayDateFormat(new Date(), DATE_DISPLAY.WEEKDAY_DAY_MONTH_YEAR),
}))

const authStore = useAuthStore()

const settings = ref<IUserSettings>({
	...authStore.settings,
	frontendSettings: {
		// Sub objects get exported as read only as well, so we need to 
		// explicitly spread the object here to allow modification
		...authStore.settings.frontendSettings,
		// Add fallback for old settings that don't have the default view set
		defaultView: authStore.settings.frontendSettings.defaultView ?? DEFAULT_PROJECT_VIEW_SETTINGS.FIRST,
		// Add fallback for old settings that don't have the minimum priority set
		minimumPriority: authStore.settings.frontendSettings.minimumPriority ?? PRIORITIES.MEDIUM,
		// Add fallback for old settings that don't have the logo change setting set
		allowIconChanges: authStore.settings.frontendSettings.allowIconChanges ?? true,
		dateDisplay: authStore.settings.frontendSettings.dateDisplay ?? DATE_DISPLAY.RELATIVE,
	},
})

const initialSettings = ref<IUserSettings>()
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

watch(
	() => authStore.settings,
	(newVal) => {
		if (Object.keys(settings.value).length !== 0) {
			return
		}
		initialSettings.value = JSON.parse(JSON.stringify({
			...newVal,
			frontendSettings: {
				...newVal.frontendSettings,
			},
		}))
		isDirty.value = !isEqual(settings.value, initialSettings.value)
	},
	{deep: true},
)

function useAvailableTimezones(settingsRef: Ref<IUserSettings>) {
	const availableTimezones = ref<{value: string, label: string}[]>([])
	const searchResults = ref<{value: string, label: string}[]>([])

	// Load timezones from API
	const HTTP = AuthenticatedHTTPFactory()
	HTTP.get('user/timezones')
		.then(r => {
			if (r.data) {
				// Transform timezones into objects with value/label pairs
				availableTimezones.value = r.data
					.sort((a, b) => a.localeCompare(b))
					.map((tz: string) => ({
						value: tz,
						label: tz.replace(/_/g, ' '),
					}))
				
				// Initial populate of search results
				searchResults.value = [...availableTimezones.value]
				return
			}
			
			availableTimezones.value = []
		})
	
	// Search function that filters available timezones
	function search(query: string) {
		if (query === '') {
			searchResults.value = [...availableTimezones.value]
			return
		}

		searchResults.value = availableTimezones.value
			.filter(tz => tz.label.toLowerCase().includes(query.toLowerCase()))
	}
	
	const timezoneObject = computed({
		get: () => ({ 
			value: settingsRef.value.timezone, 
			label: settingsRef.value.timezone?.replace(/_/g, ' '), 
		}),
		set: (obj) => {
			if (obj && typeof obj === 'object' && 'value' in obj) {
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

const id = ref(createRandomID())
const availableLanguageOptions = ref(
	Object.entries(SUPPORTED_LOCALES)
		.map(l => ({code: l[0], title: l[1]}))
		.sort((a, b) => a.title.localeCompare(b.title)),
)

const isExternalUser = computed(() => !authStore.info.isLocalUser)

watch(
	() => authStore.settings,
	() => {
		// Only set setting if we don't have edited values yet to avoid overriding
		if (Object.keys(settings.value).length !== 0) {
			return
		}
		settings.value = {...authStore.settings}
	},
	{immediate: true},
)

const projectStore = useProjectStore()
const defaultProject = computed({
	get: () => projectStore.projects[settings.value.defaultProjectId],
	set(l) {
		settings.value.defaultProjectId = l ? l.id : DEFAULT_PROJECT_ID
	},
})
const filterUsedInOverview = computed({
	get: () => projectStore.projects[settings.value.frontendSettings.filterIdUsedOnOverview],
	set(l) {
		settings.value.frontendSettings.filterIdUsedOnOverview = l ? l.id : null
	},
})
const hasFilters = computed(() => typeof projectStore.projectsArray.find(p => isSavedFilter(p)) !== 'undefined')
const loading = computed(() => authStore.isLoadingGeneralSettings)

async function updateSettings() {
	await authStore.saveUserSettings({
		settings: {...settings.value},
	})
	initialSettings.value = JSON.parse(JSON.stringify(settings.value))
	isDirty.value = false
}
</script>

<style lang="scss" scoped>
.select select {
	inline-size: 100%;
}

.timezone-select {
	min-inline-size: 200px;
	flex-grow: 1;
}

.section-block + .section-block {
	margin-block-start: 1.5rem;
}

.field-group {
	display: grid;
	grid-template-columns: 1fr;
}

.field > label.two-col {
	display: flex;
	align-items: center;
	gap: .5rem;

	> span {
		flex: 0 0 50%;
	}

	input, .input, .select, .timezone-select, :deep(.multiselect) {
		flex: 0 0 50%;
		box-sizing: border-box;
	}
}

label.checkbox {
	display: flex;
	gap: .5rem;
}

.sticky-save {
	position: sticky;
	inset-block-end: 0;
	padding: .25rem 1rem 1rem;
}
</style>
