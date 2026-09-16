<template>
	<SharingList
		v-model:search="search"
		:project-id="projectId"
		:user-is-admin="userIsAdmin"
		:shares="shares"
		:candidates="candidates"
		:search-loading="isFetching"
		search-label="username"
		:is-mutating="createShare.isPending.value || updateShare.isPending.value || deleteShare.isPending.value"
		:type-name="$t('project.share.userTeam.typeUser', 1)"
		:type-names="$t('project.share.userTeam.typeUser', 2)"
		:share-name="getDisplayName"
		:add="(projectId, user) => createShare.mutateAsync({projectId, username: user.username})"
		:update-permission="(projectId, share, permission) => updateShare.mutateAsync({projectId, username: share.username, permission})"
		:remove="(projectId, share) => deleteShare.mutateAsync({projectId, username: share.username})"
	>
		<template #candidate="{candidate}">
			<User
				:avatar-size="24"
				:show-username="true"
				:user="candidate"
			/>
		</template>
		<template #share="{share}">
			<td>{{ getDisplayName(share) }}</td>
			<td>
				<b
					v-if="share.id === authStore.info?.id"
					class="is-success"
				>{{ $t('project.share.userTeam.you') }}</b>
			</td>
		</template>
	</SharingList>
</template>

<script setup lang="ts">
import {computed, ref} from 'vue'
import {useQuery} from '@tanstack/vue-query'
import {projectUserSharesQuery, useCreateProjectUserShareMutation, useUpdateProjectUserShareMutation, useDeleteProjectUserShareMutation} from '@/client/queries/projectShares'
import {useUserSearch} from '@/composables/useUserSearch'
import {getDisplayName} from '@/models/user'
import {useAuthStore} from '@/stores/auth'
import SharingList from '@/components/sharing/SharingList.vue'
import User from '@/components/misc/User.vue'

const props = defineProps<{
	projectId: number
	userIsAdmin: boolean
}>()

const authStore = useAuthStore()

function hasUsername<T extends {username?: string}>(user: T): user is T & {username: string} {
	return Boolean(user.username)
}

const {data} = useQuery(computed(() => ({...projectUserSharesQuery(props.projectId), enabled: props.projectId > 0})))
const shares = computed(() => (data.value ?? []).filter(hasUsername))
const search = ref('')
const {users, isFetching} = useUserSearch(search)
const candidates = computed(() => users.value.filter(hasUsername).filter(user => user.id !== authStore.info?.id))

const createShare = useCreateProjectUserShareMutation()
const updateShare = useUpdateProjectUserShareMutation()
const deleteShare = useDeleteProjectUserShareMutation()
</script>
