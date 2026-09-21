<script lang="ts" setup>
import {computed} from 'vue'
import {useConfigStore} from '@/stores/config'
import {useBaseStore} from '@/stores/base'
import {useTasks} from '@/composables/useTasks'
const configStore = useConfigStore()
const baseStore = useBaseStore()
const migratorsEnabled = computed(() => configStore.available_migrators?.length > 0)
const query = useTasks({params: {per_page: 1}}, {enabled: () => migratorsEnabled.value && !baseStore.hasTasks})
const loading = query.isFetching
const show = computed(() => migratorsEnabled.value
	&& !baseStore.hasTasks
	&& query.isSuccess.value
	&& query.tasks.value.length === 0)
</script>

<template>
	<template v-if="show && !loading">
		<p class="mbs-4">
			{{ $t('home.project.importText') }}
		</p>
		<XButton
			:to="{ name: 'migrate.start' }"
			:shadow="false"
		>
			{{ $t('home.project.import') }}
		</XButton>
	</template>
</template>
