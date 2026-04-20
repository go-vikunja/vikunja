<template>
	<Card
		:title="$t('user.settings.sections.appearance')"
		class="general-settings"
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
		</div>
	</Card>

	<div class="sticky-save">
		<x-button
			:loading="loading"
			@click="updateSettings"
		>
			{{ $t('misc.save') }}
		</x-button>
	</div>
</template>

<script setup lang="ts">
import {computed, ref} from 'vue'
import {useI18n} from 'vue-i18n'

import {PrefixMode} from '@/modules/quickAddMagic'
import type {IUserSettings} from '@/modelTypes/IUserSettings'

import {useTitle} from '@/composables/useTitle'
import {useAuthStore} from '@/stores/auth'

defineOptions({name: 'UserSettingsAppearance'})

const {t} = useI18n({useScope: 'global'})
useTitle(() => `${t('user.settings.sections.appearance')} - ${t('user.settings.title')}`)

const authStore = useAuthStore()

const settings = ref<IUserSettings>({
	...authStore.settings,
	frontendSettings: {
		...authStore.settings.frontendSettings,
	},
})

const loading = computed(() => authStore.isLoadingGeneralSettings)

const colorSchemeSettings = computed(() => ({
	light: t('user.settings.appearance.colorScheme.light'),
	auto: t('user.settings.appearance.colorScheme.system'),
	dark: t('user.settings.appearance.colorScheme.dark'),
}))

async function updateSettings() {
	await authStore.saveUserSettings({
		settings: {...settings.value},
	})

	settings.value = {
		...authStore.settings,
		frontendSettings: {
			...authStore.settings.frontendSettings,
		},
	}
}
</script>

<style lang="scss" scoped>
.select select {
	inline-size: 100%;
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

	.input, .select {
		flex: 0 0 50%;
		box-sizing: border-box;
	}
}

.sticky-save {
	position: sticky;
	inset-block-end: 0;
	padding: .25rem 1rem 1rem;
}
</style>