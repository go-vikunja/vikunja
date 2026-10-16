<template>
	<div class="content is-max-width-desktop">
		<h1>{{ $t('trash.title') }}</h1>
		<nav class="trash-tabs">
			<BaseButton
				class="trash-tab"
				:class="{'is-active': route.name === 'trash.tasks'}"
				:to="{name: 'trash.tasks'}"
			>
				{{ $t('quickActions.tasks') }}
			</BaseButton>
		</nav>
		<RouterView />
	</div>
</template>

<script setup lang="ts">
import {useRoute} from 'vue-router'
import {useI18n} from 'vue-i18n'

import BaseButton from '@/components/base/BaseButton.vue'
import {useTitle} from '@/composables/useTitle'

const route = useRoute()
const {t} = useI18n({useScope: 'global'})
useTitle(() => t('trash.title'))
</script>

<style lang="scss" scoped>
.trash-tabs {
	display: inline-flex;
	gap: .5rem;
	margin-block-end: 1rem;
	padding: .5rem;
	border-radius: $radius;
	background: var(--white);
	box-shadow: var(--shadow-sm);
	font-size: .75rem;
}

.trash-tab {
	padding: .25rem .5rem;
	border-radius: $radius;

	&:hover,
	&.is-active {
		color: var(--switch-view-color);
		background: var(--switch-view-active-background);
	}

	&.is-active {
		font-weight: bold;
	}
}
</style>
