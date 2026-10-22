<template>
	<div
		class="content loader-container is-max-width-desktop"
		:class="{ 'is-loading': loading}"
	>
		<XButton
			:to="{name:'teams.create'}"
			class="is-pulled-end"
			icon="plus"
		>
			{{ $t('team.create.title') }}
		</XButton>

		<h1>{{ $t('team.title') }}</h1>
		<Card
			v-if="teams.length > 0"
			:padding="false"
			:has-content="false"
		>
			<ul class="teams">
				<li
					v-for="team in teams"
					:key="team.id"
				>
					<RouterLink :to="{name: 'teams.edit', params: {id: team.id}}">
						<p>
							{{ team.name }}
						</p>
					</RouterLink>
				</li>
			</ul>
		</Card>
		<p
			v-else-if="!isFetching"
			class="has-text-centered has-text-grey is-italic"
		>
			{{ $t('team.noTeams') }}
			<RouterLink :to="{name: 'teams.create'}">
				{{ $t('team.create.title') }}.
			</RouterLink>
		</p>
		<Pagination
			:total-pages="totalPages"
			:current-page="page"
		/>
	</div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import {useRouteQuery} from '@vueuse/router'

import Card from '@/components/misc/Card.vue'
import Pagination from '@/components/misc/Pagination.vue'
import {normalizePageNumber} from '@/client/queries/pagination'
import {useClampedPage} from '@/composables/useClampedPage'
import {useTeamsPage} from '@/composables/useTeams'
import { useTitle } from '@/composables/useTitle'
import {useDelayedLoading} from '@/composables/useDelayedLoading'

const { t } = useI18n({useScope: 'global'})
useTitle(() => t('team.title'))

const page = useRouteQuery('page', '1', {transform: normalizePageNumber})
const teamsPage = useTeamsPage(page)
const {teams, totalPages, isFetching} = teamsPage
useClampedPage(page, teamsPage)
const loading = useDelayedLoading(isFetching)
</script>

<style lang="scss" scoped>
ul.teams {
  padding: 0;
  margin-block-start: 0;
  margin-inline-start: 0;
  border-radius: $radius;
  overflow: hidden;

  li {
    list-style: none;
    margin: 0;
    border-inline-end: 1px solid var(--grey-200);

    a {
      color: var(--text);
      display: block;
      padding: 0.5rem 1rem;
      transition: background-color $transition;

      &:hover {
        background: var(--grey-100);
      }
    }
  }

  li:last-child {
    border-inline-end: none;
  }
}
</style>
