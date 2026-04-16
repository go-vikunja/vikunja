<template>
	<div class="content-widescreen">
		<div class="admin-shell">
			<nav
				class="navigation"
				data-cy="admin-shell-nav"
			>
				<ul>
					<li
						v-for="({routeName, title}, index) in navigationItems"
						:key="index"
					>
						<RouterLink
							class="navigation-link"
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
import {useI18n} from 'vue-i18n'
import {useTitle} from '@/composables/useTitle'

const {t} = useI18n({useScope: 'global'})
useTitle(() => t('admin.title'))

const navigationItems = computed(() => [
	{
		title: t('admin.nav.overview'),
		routeName: 'admin.overview',
	},
	{
		title: t('admin.nav.users'),
		routeName: 'admin.users',
	},
	{
		title: t('admin.nav.projects'),
		routeName: 'admin.projects',
	},
])
</script>

<style lang="scss" scoped>
.admin-shell {
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
