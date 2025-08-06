<template>
	<div class="content-widescreen">
		<div class="user-settings">
			<nav class="navigation">
				<ul>
					<li
						v-for="({routeName, title }, index) in navigationItems"
						:key="index"
					>
						<RouterLink
							class="navigation-link"
							:class="{'router-link-active': routeName === 'migrate.start' && route.name === 'migrate.service'}"
							:to="{name: routeName}"
						>
							{{ title }}
						</RouterLink>
					</li>
					<li
						v-for="({url, text}, index) in extraSettingsLinks"
						:key="index"
					>
						<BaseButton
							class="navigation-link is-flex is-align-items-center"
							:href="url"
						>
							<span>
								{{ text }}
							</span>
							<span class="ml-1 has-text-grey-light is-size-7">
								<Icon
									icon="arrow-up-right-from-square"
								/>
							</span>
						</BaseButton>
					</li>
				</ul>
			</nav>
			<section class="view">
				<RouterView />
			</section>
		</div>
	</div>
</template>

<script setup lang="ts">
import {computed} from 'vue'
import { useI18n } from 'vue-i18n'
import { useTitle } from '@/composables/useTitle'
import { useConfigStore } from '@/stores/config'
import { useAuthStore } from '@/stores/auth'
import {useRoute} from 'vue-router'

import BaseButton from '@/components/base/BaseButton.vue'

const { t } = useI18n({useScope: 'global'})
useTitle(() => t('user.settings.title'))

const configStore = useConfigStore()
const authStore = useAuthStore()
const route = useRoute()

const totpEnabled = computed(() => configStore.totpEnabled)
const caldavEnabled = computed(() => configStore.caldavEnabled)
const migratorsEnabled = computed(() => configStore.migratorsEnabled)
const isLocalUser = computed(() => authStore.info?.isLocalUser)
const userDeletionEnabled = computed(() => configStore.userDeletionEnabled)

const navigationItems = computed(() => {
	const items = [
		{
			title: t('user.settings.general.title'),
			routeName: 'user.settings.general',
		},
		{
			title: t('user.settings.newPasswordTitle'),
			routeName: 'user.settings.password-update',
			condition: isLocalUser.value,
		},
		{
			title: t('user.settings.updateEmailTitle'),
			routeName: 'user.settings.email-update',
			condition: isLocalUser.value,
		},
		{
			title: t('user.settings.avatar.title'),
			routeName: 'user.settings.avatar',
		},
		{
			title: t('user.settings.totp.title'),
			routeName: 'user.settings.totp',
			condition: totpEnabled.value && isLocalUser.value,
		},
		{
			title: t('user.export.title'),
			routeName: 'user.settings.data-export',
		},
		{
			title: t('migrate.title'),
			routeName: 'migrate.start',
			condition: migratorsEnabled.value,
		},
		{
			title: t('user.settings.caldav.title'),
			routeName: 'user.settings.caldav',
			condition: caldavEnabled.value,
		},
		{
			title: t('user.settings.apiTokens.title'),
			routeName: 'user.settings.apiTokens',
		},
		{
			title: t('user.deletion.title'),
			routeName: 'user.settings.deletion',
			condition: userDeletionEnabled.value,
		},
	]
	
	return items.filter(({condition}) => condition !== false)
})

const extraSettingsLinks = computed(() => authStore.settings.extraSettingsLinks)
</script>

<style lang="scss" scoped>
.user-settings {
	display: flex;

	@media screen and (max-width: $tablet) {
		flex-direction: column;
	}
}

.navigation {
	inline-size: 25%;
	padding-inline-end: 1rem;

	@media screen and (max-width: $tablet) {
		inline-size: 100%;
		padding-inline-start: 0;
	}
}

.navigation-link {
	display: block;
	padding: .5rem;
	color: var(--text);
	inline-size: 100%;
	border-inline-start: 3px solid transparent;

	&:hover,
	&.router-link-active {
		background: var(--white);
		border-color: var(--primary);
	}
}

.view {
	inline-size: 75%;

	@media screen and (max-width: $tablet) {
		inline-size: 100%;
		padding-inline-start: 0;
		padding-block-start: 1rem;
	}
}
</style>
