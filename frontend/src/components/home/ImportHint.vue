<script lang="ts" setup>
import { computed, onMounted, ref } from 'vue'
import { useConfigStore } from '@/stores/config'
import { useBaseStore } from '@/stores/base'
import { useTaskStore } from '@/stores/tasks'
import TaskService from '@/services/task'

const configStore = useConfigStore()
const baseStore = useBaseStore()
const taskStore = useTaskStore()

const migratorsEnabled = computed(() => configStore.availableMigrators?.length > 0)
const hasTasks = computed(() => baseStore.hasTasks)
const loading = computed(() => taskStore.isLoading)
const show = ref(false)

onMounted(async () => {
	show.value = false

	if (!migratorsEnabled.value) {
		return
	}

	if (hasTasks.value) {
		show.value = false
		return
	}

	const taskService = new TaskService()
	const tasks = await taskService.getAll({}, {per_page: 1})
	show.value = tasks.length === 0
})
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
