<template>
	<card :title="$t('user.settings.general.title')" class="general-settings" :loading="userSettingsService.loading">
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
					<select v-model="language">
						<option :value="lang.code" v-for="lang in availableLanguages" :key="lang.code">{{
								lang.title
							}}
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
						<option v-for="set in quickAddMagicPrefixes" :key="set" :value="set">
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

		<x-button
			:loading="userSettingsService.loading"
			@click="updateSettings()"
			class="is-fullwidth mt-4"
		>
			{{ $t('misc.save') }}
		</x-button>
	</card>
</template>

<script>
import {computed, watch} from 'vue'
import {useI18n} from 'vue-i18n'

import {playSoundWhenDoneKey} from '@/helpers/playPop'
import {availableLanguages, saveLanguage, getCurrentLanguage} from '@/i18n'
import {getQuickAddMagicMode, setQuickAddMagicMode} from '@/helpers/quickAddMagicMode'
import UserSettingsService from '@/services/userSettings'
import {PrefixMode} from '@/modules/parseTaskText'
import ListSearch from '@/components/tasks/partials/listSearch'
import {createRandomID} from '@/helpers/randomId'
import {playPop} from '@/helpers/playPop'
import {useColorScheme} from '@/composables/useColorScheme'
import {success} from '@/message'

function useColorSchemeSetting() {
	const { t } = useI18n()
	const colorSchemeSettings = computed(() => ({
		light: t('user.settings.appearance.colorScheme.light'),
		auto: t('user.settings.appearance.colorScheme.system'),
		dark: t('user.settings.appearance.colorScheme.dark'),
	}))

	const {store} = useColorScheme()
	watch(store, (schemeId) => {
		success({message: t('user.settings.appearance.setSuccess', {
			colorScheme: colorSchemeSettings.value[schemeId],
		})})
	})

	return {
		colorSchemeSettings,
		activeColorSchemeSetting: store,
	}
}

function getPlaySoundWhenDoneSetting() {
	return localStorage.getItem(playSoundWhenDoneKey) === 'true' || localStorage.getItem(playSoundWhenDoneKey) === null
}

export default {
	name: 'user-settings-general',
	data() {
		return {
			playSoundWhenDone: getPlaySoundWhenDoneSetting(),
			language: getCurrentLanguage(),
			quickAddMagicMode: getQuickAddMagicMode(),
			quickAddMagicPrefixes: PrefixMode,
			userSettingsService: new UserSettingsService(),
			settings: {...this.$store.state.auth.settings},
			id: createRandomID(),
		}
	},
	components: {
		ListSearch,
	},
	computed: {
		availableLanguages() {
			return Object.entries(availableLanguages)
				.map(l => ({code: l[0], title: l[1]}))
				.sort((a, b) => a.title.localeCompare(b.title))
		},
		defaultList() {
			return this.$store.getters['lists/getListById'](this.settings.defaultListId)
		},
	},

	setup() {
		return {
			...useColorSchemeSetting(),
		}
	},

	mounted() {
		this.setTitle(`${this.$t('user.settings.general.title')} - ${this.$t('user.settings.title')}`)
	},
	watch: {
		playSoundWhenDone(play) {
			if (play) {
				playPop()
			}
		},
	},
	methods: {
		async updateSettings() {
			localStorage.setItem(playSoundWhenDoneKey, this.playSoundWhenDone)
			saveLanguage(this.language)
			setQuickAddMagicMode(this.quickAddMagicMode)
			this.settings.defaultListId = this.defaultList ? this.defaultList.id : 0

			await this.userSettingsService.update(this.settings)
			this.$store.commit('auth/setUserSettings', {
				...this.settings,
			})
			this.$message.success({message: this.$t('user.settings.general.savedSuccess')})
		},
	},
}
</script>
