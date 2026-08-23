<template>
	<ProjectImagePicker
		i18n-prefix="project.background"
		:unsplash-service="backgroundUnsplashService"
		:upload-service="backgroundUploadService"
		:upload-enabled="uploadBackgroundEnabled"
		:unsplash-enabled="unsplashBackgroundEnabled"
		:has-image="hasBackground"
		:remove-image="removeBackground"
	/>
</template>

<script setup lang="ts">
import {computed, shallowReactive} from 'vue'

import {useConfigStore} from '@/stores/config'

import BackgroundUnsplashService from '@/services/backgroundUnsplash'
import BackgroundUploadService from '@/services/backgroundUpload'
import ProjectService from '@/services/project'
import type {IProject} from '@/modelTypes/IProject'

import ProjectImagePicker from '@/components/project/ProjectImagePicker.vue'

defineOptions({name: 'ProjectSettingBackground'})

const backgroundUnsplashService = shallowReactive(new BackgroundUnsplashService())
const backgroundUploadService = shallowReactive(new BackgroundUploadService())
const projectService = new ProjectService()
const configStore = useConfigStore()

const unsplashBackgroundEnabled = computed(() => configStore.enabledBackgroundProviders.includes('unsplash'))
const uploadBackgroundEnabled = computed(() => configStore.enabledBackgroundProviders.includes('upload'))

function hasBackground(project: IProject) {
	return !!project.backgroundInformation
}

function removeBackground(project: IProject) {
	return projectService.removeBackground(project)
}
</script>
