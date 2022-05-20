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
				<input type="checkbox" v-model="settings.emailRemindersEnabled"/>
				{{ $t('user.settings.general.emailReminders') }}
			</label>
		</div>
		<div class="field">
			<label class="checkbox">
				<input type="checkbox" v-model="settings.overdueTasksRemindersEnabled"/>
				{{ $t('user.settings.general.overdueReminders') }}
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
						>{{ lang.title }}</option>
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
import {defineComponent} from 'vue'

export default defineComponent({
	name: 'user-settings-general',
})
</script>

<script setup lang="ts">
import {computed, watch, ref, reactive} from 'vue'
import {useI18n} from 'vue-i18n'
import {useStore} from 'vuex'

import {PrefixMode} from '@/modules/parseTaskText'

import ListSearch from '@/components/tasks/partials/listSearch.vue'

import {availableLanguages} from '@/i18n'
import {playSoundWhenDoneKey, playPopSound} from '@/helpers/playPop'
import {getQuickAddMagicMode, setQuickAddMagicMode} from '@/helpers/quickAddMagicMode'
import {createRandomID} from '@/helpers/randomId'
import {success} from '@/message'
import {AuthenticatedHTTPFactory} from '@/http-common'

import {useColorScheme} from '@/composables/useColorScheme'
import { useTitle } from '@/composables/useTitle'

const {t} = useI18n()
useTitle(() => `${t('user.settings.general.title')} - ${t('user.settings.title')}`)

const DEFAULT_LIST_ID = 0

function useColorSchemeSetting() {
	const {t} = useI18n()
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

const store = useStore()
const settings = reactive({...store.state.auth.settings})
const id = ref(createRandomID())
const availableLanguageOptions = ref(
	Object.entries(availableLanguages)
		.map(l => ({code: l[0], title: l[1]}))
		.sort((a, b) => a.title.localeCompare(b.title)),
)


const defaultList = computed({
	get: () => store.getters['lists/getListById'](settings.defaultListId),
	set(l) {
		settings.defaultListId = l ? l.id : DEFAULT_LIST_ID
	},
})
const loading = computed(() => store.state.loading && store.state.loadingModule === 'general-settings')

watch(
	playSoundWhenDone,
	(play) => play && playPopSound(),
)


async function updateSettings() {
	localStorage.setItem(playSoundWhenDoneKey, playSoundWhenDone.value ? 'true' : 'false')
	setQuickAddMagicMode(quickAddMagicMode.value)

	await store.dispatch('auth/saveUserSettings', {
		settings: {...settings},
	})
}
</script>
