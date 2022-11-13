<template>
	<div class="content loader-container" :class="{'is-loading': loading}" v-cy="'namespaces-list'">
		<header class="namespace-header">
			<fancycheckbox v-model="showArchived" v-cy="'show-archived-check'">
				{{ $t('namespace.showArchived') }}
			</fancycheckbox>

			<div class="action-buttons">
				<x-button :to="{name: 'filters.create'}" icon="filter">
					{{ $t('filters.create.title') }}
				</x-button>
				<x-button :to="{name: 'namespace.create'}" icon="plus" v-cy="'new-namespace'">
					{{ $t('namespace.create.title') }}
				</x-button>
			</div>
		</header>

		<p v-if="namespaces.length === 0" class="has-text-centered has-text-grey mt-4 is-italic">
			{{ $t('namespace.noneAvailable') }}
			<BaseButton :to="{name: 'namespace.create'}">
				{{ $t('namespace.create.title') }}.
			</BaseButton>
		</p>

		<section :key="`n${n.id}`" class="namespace" v-for="n in namespaces">
			<x-button
				v-if="n.id > 0 && n.projects.length > 0"
				:to="{name: 'project.create', params: {namespaceId:  n.id}}"
				class="is-pulled-right"
				variant="secondary"
				icon="plus"
			>
				{{ $t('project.create.header') }}
			</x-button>
			<x-button
				v-if="n.isArchived"
				:to="{name: 'namespace.settings.archive', params: {id:  n.id}}"
				class="is-pulled-right mr-4"
				variant="secondary"
				icon="archive"
			>
				{{ $t('namespace.unarchive') }}
			</x-button>

			<h2 class="namespace-title">
				<span v-cy="'namespace-title'">{{ getNamespaceTitle(n) }}</span>
				<span v-if="n.isArchived" class="is-archived">
					{{ $t('namespace.archived') }}
				</span>
			</h2>

			<p v-if="n.projects.length === 0" class="has-text-centered has-text-grey mt-4 is-italic">
				{{ $t('namespace.noProjects') }}
				<BaseButton :to="{name: 'project.create', params: {namespaceId:  n.id}}">
					{{ $t('namespace.createProject') }}
				</BaseButton>
			</p>

			<ProjectCardGrid v-else 			
				:projects="n.projects"
				:show-archived="showArchived"
			/>
		</section>
	</div>
</template>

<script setup lang="ts">
import {computed} from 'vue'
import {useI18n} from 'vue-i18n'

import BaseButton from '@/components/base/BaseButton.vue'
import Fancycheckbox from '@/components/input/fancycheckbox.vue'
import ProjectCardGrid from '@/components/project/partials/ProjectCardGrid.vue'

import {getNamespaceTitle} from '@/helpers/getNamespaceTitle'
import {useTitle} from '@/composables/useTitle'
import {useStorage} from '@vueuse/core'

import {useNamespaceStore} from '@/stores/namespaces'

const {t} = useI18n()
const namespaceStore = useNamespaceStore()

useTitle(() => t('namespace.title'))
const showArchived = useStorage('showArchived', false)

const loading = computed(() => namespaceStore.isLoading)
const namespaces = computed(() => {
	return namespaceStore.namespaces.filter(namespace => showArchived.value
		? true
		: !namespace.isArchived,
	)
})
</script>

<style lang="scss" scoped>
.namespace-header {
	display: flex;
	justify-content: space-between;
	align-items: center;
	gap: 1rem;

	@media screen and (max-width: $tablet) {
		flex-direction: column;
	}
}

.action-buttons {
	display: flex;
	justify-content: space-between;
	gap: 1rem;

	@media screen and (max-width: $tablet) {
		width: 100%;
		flex-direction: column;
		align-items: stretch;
	}
}

.namespace:not(:first-child) {
	margin-top: 1rem;
}

.namespace-title {
	display: flex;
	align-items: center;
}

.is-archived {
	font-size: 0.75rem;
	border: 1px solid var(--grey-500);
	color: $grey !important;
	padding: 2px 4px;
	border-radius: 3px;
	font-family: $vikunja-font;
	background: var(--white-translucent);
	margin-left: .5rem;
}
</style>