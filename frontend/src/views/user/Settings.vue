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
</script>

<style lang="scss" scoped>
.user-settings {
	display: flex;

	@media screen and (max-width: $tablet) {
		flex-direction: column;
	}
}

.navigation {
	width: 25%;
	padding-right: 1rem;

	@media screen and (max-width: $tablet) {
		width: 100%;
		padding-left: 0;
	}
}

.navigation-link {
	display: block;
	padding: .5rem;
	color: var(--text);
	width: 100%;
	border-left: 3px solid transparent;

	&:hover,
	&.router-link-active {
		background: var(--white);
		border-color: var(--primary);
	}
}

.view {
	width: 75%;

	@media screen and (max-width: $tablet) {
		width: 100%;
		padding-left: 0;
		padding-top: 1rem;
	}
}
</style>
