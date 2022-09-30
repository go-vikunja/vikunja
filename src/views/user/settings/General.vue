<template>
	<card :title="$t('user.settings.general.title')" class="general-settings" :loading="loading">
		<div class="field">
			<label class="label" :for="`newName${id}`">{{ $t('user.settings.general.name') }}</label>
			<div class="control">
				<input
					@keyup.enter="updateSettings"
					class="input"
					:id="`newName${id}`"
					:placeholder="$t('user.settings.general.newName')"
					type="text"
					v-model="settings.name"/>
			</div>
		</div>
		<div class="field">
			<label class="label">
				{{ $t('user.settings.general.defaultList') }}
			</label>
			<list-search v-model="defaultList"/>
		</div>
		<div class="field">
			<label class="checkbox">
				<input type="checkbox" v-model="settings.overdueTasksRemindersEnabled"/>
				{{ $t('user.settings.general.overdueReminders') }}
			</label>
		</div>
		<div class="field" v-if="settings.overdueTasksRemindersEnabled">
			<label class="label" for="overdueTasksReminderTime">
				{{ $t('user.settings.general.overdueTasksRemindersTime') }}
			</label>
			<div class="control">
				<input
					@keyup.enter="updateSettings"
					class="input"
					id="overdueTasksReminderTime"
					type="time"
					v-model="settings.overdueTasksRemindersTime"/>
			</div>
		</div>
		<div class="field">
			<label class="checkbox">
				<input type="checkbox" v-model="settings.emailRemindersEnabled"/>
				{{ $t('user.settings.general.emailReminders') }}
			</label>
		</div>
		<div class="field">
			<label class="checkbox">
				<input type="checkbox" v-model="settings.discoverableByName"/>
				{{ $t('user.settings.general.discoverableByName') }}
			</label>
		</div>
		<div class="field">
			<label class="checkbox">
				<input type="checkbox" v-model="settings.discoverableByEmail"/>
				{{ $t('user.settings.general.discoverableByEmail') }}
			</label>
		</div>
		<div class="field">
			<label class="checkbox">
				<input type="checkbox" v-model="playSoundWhenDone"/>
				{{ $t('user.settings.general.playSoundWhenDone') }}
			</label>
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
					<select v-model="quickAddMagicMode">
						<option v-for="set in PrefixMode" :key="set" :value="set">
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
					<select v-model="activeColorSchemeSetting">
						<!-- TODO: use the Vikunja logo in color scheme as option buttons -->
						<option v-for="(title, schemeId) in colorSchemeSettings" :key="schemeId" :value="schemeId">
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
						<option v-for="tz in availableTimezones" :key="tz">
							{{ tz }}
						</option>
					</select>
				</div>
			</label>
		</div>

		<x-button
			:loading="loading"
			@click="updateSettings()"
			class="is-fullwidth mt-4"
			v-cy="'saveGeneralSettings'"
		>
			{{ $t('misc.save') }}
		</x-button>
	</card>
</template>

<script lang="ts">
export default {name: 'user-settings-general'}
</script>

<script setup lang="ts">
import {computed, watch, ref} from 'vue'
import {useI18n} from 'vue-i18n'

import {PrefixMode} from '@/modules/parseTaskText'

import ListSearch from '@/components/tasks/partials/listSearch.vue'

import {availableLanguages} from '@/i18n'
import {playSoundWhenDoneKey, playPopSound} from '@/helpers/playPop'
import {getQuickAddMagicMode, setQuickAddMagicMode} from '@/helpers/quickAddMagicMode'
import {createRandomID} from '@/helpers/randomId'
import {objectIsEmpty} from '@/helpers/objectIsEmpty'
import {success} from '@/message'
import {AuthenticatedHTTPFactory} from '@/http-common'

import {useColorScheme} from '@/composables/useColorScheme'
import {useTitle} from '@/composables/useTitle'

import {useListStore} from '@/stores/lists'
import {useAuthStore} from '@/stores/auth'

const {t} = useI18n({useScope: 'global'})
useTitle(() => `${t('user.settings.general.title')} - ${t('user.settings.title')}`)

const DEFAULT_LIST_ID = 0

function useColorSchemeSetting() {
	const {t} = useI18n({useScope: 'global'})
	const colorSchemeSettings = computed(() => ({
		light: t('user.settings.appearance.colorScheme.light'),
		auto: t('user.settings.appearance.colorScheme.system'),
		dark: t('user.settings.appearance.colorScheme.dark'),
	}))

	const {store} = useColorScheme()
	watch(store, (schemeId) => {
		success({
			message: t('user.settings.appearance.setSuccess', {
				colorScheme: colorSchemeSettings.value[schemeId],
			}),
		})
	})

	return {
		colorSchemeSettings,
		activeColorSchemeSetting: store,
	}
}

const {colorSchemeSettings, activeColorSchemeSetting} = useColorSchemeSetting()

function useAvailableTimezones() {
	const availableTimezones = ref([])

	const HTTP = AuthenticatedHTTPFactory()
	HTTP.get('user/timezones')
		.then(r => {
			availableTimezones.value = r.data.sort()
		})

	return availableTimezones
}

const availableTimezones = useAvailableTimezones()

function getPlaySoundWhenDoneSetting() {
	return localStorage.getItem(playSoundWhenDoneKey) === 'true' || localStorage.getItem(playSoundWhenDoneKey) === null
}

const playSoundWhenDone = ref(getPlaySoundWhenDoneSetting())
const quickAddMagicMode = ref(getQuickAddMagicMode())

const authStore = useAuthStore()
const settings = ref({...authStore.settings})
const id = ref(createRandomID())
const availableLanguageOptions = ref(
	Object.entries(availableLanguages)
		.map(l => ({code: l[0], title: l[1]}))
		.sort((a, b) => a.title.localeCompare(b.title)),
)

watch(
	() => authStore.settings,
	() => {
		// Only setting if we don't have values set yet to avoid overriding edited values
		if (!objectIsEmpty(settings.value)) {
			return
		}
		settings.value = {...authStore.settings}
	},
	{immediate: true},
)

const listStore = useListStore()
const defaultList = computed({
	get: () => listStore.getListById(settings.value.defaultListId),
	set(l) {
		settings.value.defaultListId = l ? l.id : DEFAULT_LIST_ID
	},
})
const loading = computed(() => authStore.isLoadingGeneralSettings)

watch(
	playSoundWhenDone,
	(play) => play && playPopSound(),
)

async function updateSettings() {
	localStorage.setItem(playSoundWhenDoneKey, playSoundWhenDone.value ? 'true' : 'false')
	setQuickAddMagicMode(quickAddMagicMode.value)

	await authStore.saveUserSettings({
		settings: {...settings.value},
	})
}
</script>
