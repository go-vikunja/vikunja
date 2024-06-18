<template>
	<Card
		:title="$t('user.settings.general.title')"
		class="general-settings"
		:loading="loading"
	>
		<div class="field">
			<label
				class="label"
				:for="`newName${id}`"
			>{{ $t('user.settings.general.name') }}</label>
			<div class="control">
				<input
					:id="`newName${id}`"
					v-model="settings.name"
					class="input"
					:placeholder="$t('user.settings.general.newName')"
					type="text"
					@keyup.enter="updateSettings"
				>
			</div>
		</div>
		<div class="field">
			<label class="label">
				{{ $t('user.settings.general.defaultProject') }}
			</label>
			<ProjectSearch v-model="defaultProject" />
		</div>
		<div class="field">
			<label class="label">
				{{ $t('user.settings.general.defaultView') }}
			</label>
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
		</div>
		<div
			v-if="hasFilters"
			class="field"
		>
			<label class="label">
				{{ $t('user.settings.general.filterUsedOnOverview') }}
			</label>
			<ProjectSearch
				v-model="filterUsedInOverview"
				:saved-filters-only="true"
			/>
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
				class="label"
				for="overdueTasksReminderTime"
			>
				{{ $t('user.settings.general.overdueTasksRemindersTime') }}
			</label>
			<div class="control">
				<input
					id="overdueTasksReminderTime"
					v-model="settings.overdueTasksRemindersTime"
					class="input"
					type="time"
					@keyup.enter="updateSettings"
				>
			</div>
		</div>
		<div class="field">
			<label class="is-flex is-align-items-center">
				<span>
					{{ $t('user.settings.general.weekStart') }}
				</span>
				<div class="select ml-2">
					<select v-model.number="settings.weekStart">
						<option value="0">{{ $t('user.settings.general.weekStartSunday') }}</option>
						<option value="1">{{ $t('user.settings.general.weekStartMonday') }}</option>
					</select>
				</div>
			</label>
		</div>
		<div class="field">
			<label class="is-flex is-align-items-center">
				<span>
					{{ $t('user.settings.general.language') }}
				</span>
				<div class="select ml-2">
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
			<label class="is-flex is-align-items-center">
				<span>
					{{ $t('user.settings.quickAddMagic.title') }}
				</span>
				<div class="select ml-2">
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
			<label class="is-flex is-align-items-center">
				<span>
					{{ $t('user.settings.appearance.title') }}
				</span>
				<div class="select ml-2">
					<select v-model="settings.frontendSettings.colorSchema">
						<!-- TODO: use the Vikunja logo in color scheme as option buttons -->
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
			<label class="is-flex is-align-items-center">
				<span>
					{{ $t('user.settings.general.timezone') }}
				</span>
				<div class="select ml-2">
					<select v-model="settings.timezone">
						<option
							v-for="tz in availableTimezones"
							:key="tz"
						>
							{{ tz }}
						</option>
					</select>
				</div>
			</label>
		</div>

		<x-button
			v-cy="'saveGeneralSettings'"
			:loading="loading"
			class="is-fullwidth mt-4"
			@click="updateSettings()"
		>
			{{ $t('misc.save') }}
		</x-button>
	</Card>
</template>

<script lang="ts">
export default {name: 'UserSettingsGeneral'}
</script>

<script setup lang="ts">
import {computed, watch, ref} from 'vue'
import {useI18n} from 'vue-i18n'

import {PrefixMode} from '@/modules/parseTaskText'

import ProjectSearch from '@/components/tasks/partials/ProjectSearch.vue'

import {SUPPORTED_LOCALES} from '@/i18n'
import {createRandomID} from '@/helpers/randomId'
import {AuthenticatedHTTPFactory} from '@/helpers/fetcher'

import {useTitle} from '@/composables/useTitle'

import {useProjectStore} from '@/stores/projects'
import {useAuthStore} from '@/stores/auth'
import type {IUserSettings} from '@/modelTypes/IUserSettings'
import {isSavedFilter} from '@/services/savedFilter'
import {DEFAULT_PROJECT_VIEW_SETTINGS} from '@/modelTypes/IProjectView'

const {t} = useI18n({useScope: 'global'})
useTitle(() => `${t('user.settings.general.title')} - ${t('user.settings.title')}`)

const DEFAULT_PROJECT_ID = 0

const colorSchemeSettings = computed(() => ({
	light: t('user.settings.appearance.colorScheme.light'),
	auto: t('user.settings.appearance.colorScheme.system'),
	dark: t('user.settings.appearance.colorScheme.dark'),
}))

function useAvailableTimezones() {
	const availableTimezones = ref([])

	const HTTP = AuthenticatedHTTPFactory()
	HTTP.get('user/timezones')
		.then(r => {
			if (r.data) {
				availableTimezones.value = r.data.sort()
				return
			}
			
			availableTimezones.value = []
		})

	return availableTimezones
}

const availableTimezones = useAvailableTimezones()

const authStore = useAuthStore()

const settings = ref<IUserSettings>({
	...authStore.settings,
	frontendSettings: {
		// Sub objects get exported as read only as well, so we need to 
		// explicitly spread the object here to allow modification
		...authStore.settings.frontendSettings,
		// Add fallback for old settings that don't have the default view set
		defaultView: authStore.settings.frontendSettings.defaultView ?? DEFAULT_PROJECT_VIEW_SETTINGS.FIRST,
	},
})
const id = ref(createRandomID())
const availableLanguageOptions = ref(
	Object.entries(SUPPORTED_LOCALES)
		.map(l => ({code: l[0], title: l[1]}))
		.sort((a, b) => a.title.localeCompare(b.title)),
)

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
}
</script>
