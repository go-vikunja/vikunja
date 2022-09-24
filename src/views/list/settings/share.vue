<template>
	<create-edit
		:title="$t('list.share.header')"
		primary-label=""
	>
		<template v-if="list">
			<userTeam
				:id="list.id"
				:userIsAdmin="userIsAdmin"
				shareType="user"
				type="list"
			/>
			<userTeam
				:id="list.id"
				:userIsAdmin="userIsAdmin"
				shareType="team"
				type="list"
			/>
		</template>

		<link-sharing :list-id="listId" v-if="linkSharingEnabled" class="mt-4"/>
	</create-edit>
</template>

<script lang="ts">
export default {name: 'list-setting-share'}
</script>

<script lang="ts" setup>
import {ref, computed, watchEffect} from 'vue'
import {useRoute} from 'vue-router'
import {useI18n} from 'vue-i18n'
import {useTitle} from '@vueuse/core'

import ListService from '@/services/list'
import ListModel from '@/models/list'

import CreateEdit from '@/components/misc/create-edit.vue'
import LinkSharing from '@/components/sharing/linkSharing.vue'
import userTeam from '@/components/sharing/userTeam.vue'

import {useBaseStore} from '@/stores/base'
import {useConfigStore} from '@/stores/config'
import {useAuthStore} from '@/stores/auth'

const {t} = useI18n({useScope: 'global'})

const list = ref()
const title = computed(() => list.value?.title
	? t('list.share.title', {list: list.value.title})
	: '',
)
useTitle(title)

const authStore = useAuthStore()
const configStore = useConfigStore()

const linkSharingEnabled = computed(() => configStore.linkSharingEnabled)
const userIsAdmin = computed(() => 'owner' in list.value && list.value.owner.id === authStore.info.id)

async function loadList(listId: number) {
	const listService = new ListService()
	const newList = await listService.get(new ListModel({id: listId}))
	await useBaseStore().handleSetCurrentList({list: newList})
	list.value = newList
}

const route = useRoute()
const listId = computed(() => route.params.listId !== undefined
	? parseInt(route.params.listId as string)
	: undefined,
)
watchEffect(() => listId.value !== undefined && loadList(listId.value))
</script>
