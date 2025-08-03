<template>
	<div
		class="content loader-container is-max-width-desktop"
		:class="{ 'is-loading': teamService.loading}"
	>
		<XButton
			:to="{name:'teams.create'}"
			class="is-pulled-right"
			icon="plus"
		>
			{{ $t('team.create.title') }}
		</XButton>

		<h1>{{ $t('team.title') }}</h1>
		<ul
			v-if="teams.length > 0"
			class="teams box"
		>
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
		<p
			v-else-if="!teamService.loading"
			class="has-text-centered has-text-grey is-italic"
		>
			{{ $t('team.noTeams') }}
			<RouterLink :to="{name: 'teams.create'}">
				{{ $t('team.create.title') }}.
			</RouterLink>
		</p>
	</div>
</template>

<script setup lang="ts">
import {ref, shallowReactive} from 'vue'
import { useI18n } from 'vue-i18n'

import TeamService from '@/services/team'
import { useTitle } from '@/composables/useTitle'

const { t } = useI18n({useScope: 'global'})
useTitle(() => t('team.title'))

const teams = ref([])
const teamService = shallowReactive(new TeamService())
teamService.getAll().then((result) => {
	teams.value = result
})
</script>

<style lang="scss" scoped>
ul.teams {
  padding: 0;
  margin-inline-start: 0;
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
