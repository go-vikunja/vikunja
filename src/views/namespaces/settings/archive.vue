<template>
	<modal
		@close="$router.back()"
		@submit="archiveNamespace()"
	>
		<template #header><span>{{ title }}</span></template>

		<template #text>
			<p>
				{{
					namespace.isArchived
						? $t('namespace.archive.unarchiveText')
						: $t('namespace.archive.archiveText')
				}}
			</p>
		</template>
	</modal>
</template>

<script lang="ts">
export default { name: 'namespace-setting-archive' }
</script>

<script setup lang="ts">
import {watch, reactive, ref, shallowReactive} from 'vue'
import {useRouter} from 'vue-router'
import {useStore} from '@/store'
import {useI18n} from 'vue-i18n'

import {success} from '@/message'
import {useTitle} from '@/composables/useTitle'

import NamespaceService from '@/services/namespace'
import type {INamespace} from '@/modelTypes/INamespace'
import NamespaceModel from '@/models/namespace'

const props = defineProps({
	namespaceId: {
		type: Number,
		required: true,
	},
})

const store = useStore()
const router = useRouter()
const {t} = useI18n({useScope: 'global'})

const title = ref('')
useTitle(title)

const namespaceService = shallowReactive(new NamespaceService())
const namespace : INamespace = reactive(new NamespaceModel())

watch(
	() => props.namespaceId,
	async () => {
		Object.assign(namespace, store.getters['namespaces/getNamespaceById'](props.namespaceId))

		// FIXME: ressouce should be loaded in store
		Object.assign(namespace, await namespaceService.get({id: props.namespaceId}))
		title.value = namespace.isArchived ?
				t('namespace.archive.titleUnarchive', {namespace: namespace.title}) :
				t('namespace.archive.titleArchive', {namespace: namespace.title})
	},
	{ immediate: true },
)

async function archiveNamespace() {
	try {
		const isArchived = !namespace.isArchived
		const archivedNamespace = await namespaceService.update({
			...namespace,
			isArchived,
		})
		store.commit('namespaces/setNamespaceById', archivedNamespace)
		success({
			message: isArchived
				? t('namespace.archive.success')
				: t('namespace.archive.unarchiveSuccess'),
		})
	} finally {
		router.back()
	}
}
</script>
