<template>
	<div class="content is-max-width-desktop">
		<h1>{{ $t('trash.title') }}</h1>
		<div class="trash-header">
			<SwitchView
				is="nav"
				:aria-label="$t('trash.tabs.label')"
			>
				<BaseButton
					class="switch-view-button"
					:class="{'is-active': route.name === 'trash.tasks'}"
					:to="{name: 'trash.tasks'}"
				>
					{{ $t('quickActions.tasks') }}
				</BaseButton>
			</SwitchView>
			<!-- TrashTasks teleports its actions here -->
			<div id="trash-actions" />
		</div>
		<RouterView />
	</div>
</template>

<script setup lang="ts">
import {useRoute} from 'vue-router'
import {useI18n} from 'vue-i18n'

import BaseButton from '@/components/base/BaseButton.vue'
import SwitchView from '@/components/misc/SwitchView.vue'
import {useTitle} from '@/composables/useTitle'

const route = useRoute()
const {t} = useI18n({useScope: 'global'})
useTitle(() => t('trash.title'))
</script>

<style lang="scss" scoped>
.trash-header {
	display: flex;
	flex-wrap: wrap;
	align-items: center;
	justify-content: space-between;
	gap: 1rem;
	margin-block-end: 1rem;
}
</style>
