<template>
	<Modal
		@close="$router.back()"
		@submit="remove()"
	>
		<template #header>
			<span>{{ $t('filters.delete.header') }}</span>
		</template>

		<template #text>
			<ErrorMessage v-if="!(filterId > 0)" />
			<p v-else>
				{{ $t('filters.delete.text') }}
			</p>
		</template>
	</Modal>
</template>

<script setup lang="ts">
import {computed, onUnmounted, ref} from 'vue'
import {useI18n} from 'vue-i18n'
import {useRouter} from 'vue-router'

import ErrorMessage from '@/components/misc/Error.vue'

import {getSavedFilterIdFromProjectId} from '@/client/queries/projects'
import {deleteSavedFilter} from '@/client/queries/savedFilters'
import {success} from '@/message'

const props = defineProps<{
	projectId: number,
}>()

const {t} = useI18n({useScope: 'global'})
const router = useRouter()

const filterId = computed(() => getSavedFilterIdFromProjectId(props.projectId))

// useMounted() never resets on unmount.
const alive = ref(true)
onUnmounted(() => {
	alive.value = false
})

async function remove() {
	const id = props.projectId
	if (!(filterId.value > 0)) {
		return
	}

	await deleteSavedFilter(filterId.value)
	// The route param can change on this same instance, so a stale delete must not navigate.
	if (!alive.value || props.projectId !== id) {
		return
	}
	success({message: t('filters.delete.success')})
	await router.push({name: 'projects.index'})
}
</script>
