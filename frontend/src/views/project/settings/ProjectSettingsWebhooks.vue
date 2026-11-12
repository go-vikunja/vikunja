<script lang="ts" setup>
import {computed, watchEffect} from 'vue'
import {useRoute} from 'vue-router'
import {useI18n} from 'vue-i18n'
import {useTitle} from '@vueuse/core'
import {useQuery} from '@tanstack/vue-query'
import CreateEdit from '@/components/misc/CreateEdit.vue'
import WebhookManager from '@/components/misc/WebhookManager.vue'
import {useBaseStore} from '@/stores/base'
import {projectQuery} from '@/client/queries/projects'

defineOptions({name: 'ProjectSettingWebhooks'})
const {t} = useI18n({useScope: 'global'})
useTitle(t('project.webhooks.title'))
const route = useRoute()
const baseStore = useBaseStore()
const projectId = computed(() => Number(route.params.projectId))
const {data: project} = useQuery(computed(() => ({...projectQuery(projectId.value), enabled: projectId.value > 0})))
watchEffect(() => {
	if (project.value?.id === projectId.value) baseStore.setCurrentProject(project.value)
})
</script>

<template>
	<CreateEdit
		:title="$t('project.webhooks.title')"
		:has-primary-action="false"
		:wide="true"
	>
		<WebhookManager
			v-if="projectId > 0"
			:key="projectId"
			:scope="{kind: 'project', projectId}"
		/>
	</CreateEdit>
</template>
