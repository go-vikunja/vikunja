<template>
	<div class="content-widescreen">
		<div class="user-settings">
			<nav class="navigation">
				<ul>
					<li v-for="({routeName, title }, index) in navigationItems" :key="index">
						<router-link :to="{name: routeName}">
							{{ title }}
						</router-link>
					</li>
				</ul>
			</nav>
			<section class="view">
				<router-view/>
			</section>
		</div>
	</div>
</template>

<script setup lang="ts">
import {computed} from 'vue'
import { store } from '@/store'
import { useI18n } from 'vue-i18n'
import { useTitle } from '@/composables/useTitle'

const { t } = useI18n()
useTitle(() => t('user.settings.title'))

const totpEnabled = computed(() => store.state.config.totpEnabled)
const caldavEnabled = computed(() => store.state.config.caldavEnabled)
const migratorsEnabled = computed(() => store.getters['config/migratorsEnabled'])
const isLocalUser = computed(() => store.state.auth.info?.isLocalUser)

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
			condition: totpEnabled.value,
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
			title: t('user.deletion.title'),
			routeName: 'user.settings.deletion',
		},
	]
	
	return items.filter(({condition}) => condition !== false)
})
</script>

<style lang="scss" scoped>
.user-settings {
	display: flex;

	.navigation {
		width: 25%;
		padding-right: 1rem;

		a {
			display: block;
			padding: .5rem;
			color: var(--text);
			width: 100%;
			border-left: 3px solid transparent;

			&:hover, &.router-link-active {
				background: var(--white);
				border-color: var(--primary);
			}
		}
	}

	.view {
		width: 75%;
	}

	@media screen and (max-width: $tablet) {
		flex-direction: column;

		.navigation, .view {
			width: 100%;
			padding-left: 0;
		}

		.view {
			padding-top: 1rem;
		}
	}
}
</style>
