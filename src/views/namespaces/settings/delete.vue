<template>
	<modal
		@close="$router.back()"
		@submit="deleteNamespace()"
	>
		<template #header><span>{{ title }}</span></template>
		
		<template #text>
			<p>{{ $t('namespace.delete.text1') }}<br/>
			{{ $t('namespace.delete.text2') }}</p>
		</template>
	</modal>
</template>

<script lang="ts">
export default { name: 'namespace-setting-delete' }
</script>

<script setup lang="ts">
import {ref, computed, watch, shallowReactive} from 'vue'
import {useI18n} from 'vue-i18n'
import {useRouter} from 'vue-router'

import {useTitle} from '@/composables/useTitle'
import {success} from '@/message'
import {useNamespaceStore} from '@/stores/namespaces'
import NamespaceModel from '@/models/namespace'
import NamespaceService from '@/services/namespace'
import type { INamespace } from '@/modelTypes/INamespace'

const props = defineProps({
	namespaceId: {
		type: Number,
		required: true,
	},
})

const {t} = useI18n({useScope: 'global'})
const router = useRouter()
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
	return t('namespace.delete.title', {namespace: namespace.value.title})
})
useTitle(title)

async function deleteNamespace() {
	await namespaceStore.deleteNamespace(namespace.value)
	success({message: t('namespace.delete.success')})
	router.push({name: 'home'})
}
</script>
