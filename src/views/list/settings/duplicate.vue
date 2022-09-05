<template>
	<create-edit
		:title="$t('list.duplicate.title')"
		primary-icon="paste"
		:primary-label="$t('list.duplicate.label')"
		@primary="duplicateList"
		:loading="listDuplicateService.loading"
	>
		<p>{{ $t('list.duplicate.text') }}</p>

		<Multiselect
			:placeholder="$t('namespace.search')"
			@search="findNamespaces"
			:search-results="namespaces"
			@select="selectNamespace"
			label="title"
			:search-delay="10"
		/>
	</create-edit>
</template>

<script setup lang="ts">
import {ref, shallowReactive} from 'vue'
import {useRoute, useRouter} from 'vue-router'
import {useStore} from '@/store'
import {useI18n} from 'vue-i18n'

import ListDuplicateService from '@/services/listDuplicateService'
import CreateEdit from '@/components/misc/create-edit.vue'
import Multiselect from '@/components/input/multiselect.vue'

import ListDuplicateModel from '@/models/listDuplicateModel'
import type {INamespace} from '@/modelTypes/INamespace'

import {success} from '@/message'
import {useTitle} from '@/composables/useTitle'
import {useNameSpaceSearch} from '@/composables/useNamespaceSearch'

const {t} = useI18n({useScope: 'global'})
useTitle(() => t('list.duplicate.title'))

const {
	namespaces,
	findNamespaces,
} = useNameSpaceSearch()

const selectedNamespace = ref<INamespace>()

function selectNamespace(namespace: INamespace) {
	selectedNamespace.value = namespace
}

const route = useRoute()
const router = useRouter()
const store = useStore()

const listDuplicateService = shallowReactive(new ListDuplicateService())

async function duplicateList() {
	const listDuplicate = new ListDuplicateModel({
		// FIXME: should be parameter
		listId: route.params.listId,
		namespaceId: selectedNamespace.value?.id,
	})

	const duplicate = await listDuplicateService.create(listDuplicate)

	store.commit('namespaces/addListToNamespace', duplicate.list)
	store.commit('lists/setList', duplicate.list)
	success({message: t('list.duplicate.success')})
	router.push({name: 'list.index', params: {listId: duplicate.list.id}})
}
</script>
