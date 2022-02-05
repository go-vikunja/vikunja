<template>
	<create-edit
		:title="title"
		primary-label=""
	>
		<template v-if="namespace">
			<manageSharing
				:id="namespace.id"
				:userIsAdmin="userIsAdmin"
				shareType="user"
				type="namespace"
			/>
			<manageSharing
				:id="namespace.id"
				:userIsAdmin="userIsAdmin"
				shareType="team"
				type="namespace"
			/>
		</template>
	</create-edit>
</template>

<script lang="ts">
export default {
	name: 'namespace-setting-share',
}
</script>

<script lang="ts" setup>
import {ref, computed, watchEffect} from 'vue'
import {useStore} from 'vuex'
import {useRoute} from 'vue-router'
import {useI18n} from 'vue-i18n'
import {useTitle} from '@vueuse/core'

import NamespaceService from '@/services/namespace'
import NamespaceModel from '@/models/namespace'

import CreateEdit from '@/components/misc/create-edit.vue'
import manageSharing from '@/components/sharing/userTeam.vue'

const {t} = useI18n()

const namespace = ref()

const title = computed(() => namespace.value?.title
	? t('namespace.share.title', { namespace: namespace.value.title })
	: '',
)
useTitle(title)

const store = useStore()
const userIsAdmin = computed(() => 'owner' in namespace.value && namespace.value.owner.id === store.state.auth.info.id)

async function loadNamespace(namespaceId: number) {
	if (!namespaceId) return
	const namespaceService = new NamespaceService()
	namespace.value = await namespaceService.get(new NamespaceModel({id: namespaceId}))

	// TODO: set namespace in store
}

const route = useRoute()
const namespaceId = computed(() => route.params.namespaceId !== undefined
	? parseInt(route.params.namespaceId as string)
	: undefined,
)
watchEffect(() => namespaceId.value !== undefined && loadNamespace(namespaceId.value))
</script>