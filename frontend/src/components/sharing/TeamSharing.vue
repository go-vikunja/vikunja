<template>
	<SharingList
		v-model:search="search"
		:project-id="projectId"
		:user-is-admin="userIsAdmin"
		:shares="data ?? []"
		:candidates="teams"
		:search-loading="isFetching"
		search-label="name"
		:is-mutating="createShare.isPending.value || updateShare.isPending.value || deleteShare.isPending.value"
		:type-name="$t('project.share.userTeam.typeTeam', 1)"
		:type-names="$t('project.share.userTeam.typeTeam', 2)"
		:share-name="team => team.name ?? ''"
		:add="(projectId, team) => createShare.mutateAsync({projectId, teamId: team.id})"
		:update-permission="(projectId, share, permission) => updateShare.mutateAsync({projectId, teamId: share.id, permission})"
		:remove="(projectId, share) => deleteShare.mutateAsync({projectId, teamId: share.id})"
	>
		<template #candidate="{candidate}">
			<span class="search-result">{{ candidate.name }}</span>
		</template>
		<template #share="{share}">
			<td>
				<RouterLink :to="{name: 'teams.edit', params: {id: share.id}}">
					{{ share.name }}
				</RouterLink>
			</td>
		</template>
	</SharingList>
</template>

<script setup lang="ts">
import {computed, ref} from 'vue'
import {useQuery} from '@tanstack/vue-query'
import {projectTeamSharesQuery, useCreateProjectTeamShareMutation, useUpdateProjectTeamShareMutation, useDeleteProjectTeamShareMutation} from '@/client/queries/projectShares'
import {useTeams} from '@/composables/useTeams'
import {useConfigStore} from '@/stores/config'
import SharingList from '@/components/sharing/SharingList.vue'

const props = defineProps<{
	projectId: number
	userIsAdmin: boolean
}>()

const configStore = useConfigStore()

const {data} = useQuery(computed(() => ({...projectTeamSharesQuery(props.projectId), enabled: props.projectId > 0})))
const search = ref('')
const {teams, isFetching} = useTeams({search, includePublic: () => configStore.publicTeamsEnabled, enabled: () => search.value !== ''})

const createShare = useCreateProjectTeamShareMutation()
const updateShare = useUpdateProjectTeamShareMutation()
const deleteShare = useDeleteProjectTeamShareMutation()
</script>
