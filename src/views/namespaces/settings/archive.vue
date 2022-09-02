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
import {watch, ref, computed, shallowReactive, type PropType} from 'vue'
import {useRouter} from 'vue-router'
import {useI18n} from 'vue-i18n'

import {success} from '@/message'
import {useTitle} from '@/composables/useTitle'

import {useNamespaceStore} from '@/stores/namespaces'
import NamespaceService from '@/services/namespace'
import NamespaceModel from '@/models/namespace'
import type {INamespace} from '@/modelTypes/INamespace'

const props = defineProps({
	namespaceId: {
		type: Number as PropType<INamespace['id']>,
		required: true,
	},
})

const router = useRouter()
const {t} = useI18n({useScope: 'global'})

const namespaceStore = useNamespaceStore()
const namespaceService = shallowReactive(new NamespaceService())
const namespace = ref<INamespace>(new NamespaceModel())

watch(
	() => props.namespaceId,
	async () => {
		namespace.value = namespaceStore.getNamespaceById(props.namespaceId) || new NamespaceModel()

		// FIXME: ressouce should be loaded in store
		namespace.value = await namespaceService.get({id: props.namespaceId})
	},
	{ immediate: true },
)

const title = computed(() => {
	if (!namespace.value) {
		return
	}
	return namespace.value.isArchived
		? t('namespace.archive.titleUnarchive', {namespace: namespace.value.title})
		: t('namespace.archive.titleArchive', {namespace: namespace.value.title})
})
useTitle(title)

async function archiveNamespace() {
	try {
		const isArchived = !namespace.value.isArchived
		const archivedNamespace = await namespaceService.update({
			...namespace.value,
			isArchived,
		})
		namespaceStore.setNamespaceById(archivedNamespace)
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
