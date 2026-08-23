<template>
	<ProjectImagePicker
		i18n-prefix="project.cardImage"
		:unsplash-service="cardBackgroundUnsplashService"
		:upload-service="cardBackgroundUploadService"
		:upload-enabled="uploadCardImageEnabled"
		:unsplash-enabled="unsplashCardImageEnabled"
		:has-image="hasCardImage"
		:remove-image="removeCardImage"
		aspect-ratio="1 / 1"
	/>
</template>

<script setup lang="ts">
import {computed, shallowReactive} from 'vue'

import {useConfigStore} from '@/stores/config'

import CardBackgroundUnsplashService from '@/services/cardBackgroundUnsplash'
import CardBackgroundUploadService from '@/services/cardBackgroundUpload'
import ProjectService from '@/services/project'
import type {IProject} from '@/modelTypes/IProject'

import ProjectImagePicker from '@/components/project/ProjectImagePicker.vue'

defineOptions({name: 'ProjectSettingCardImage'})

const cardBackgroundUnsplashService = shallowReactive(new CardBackgroundUnsplashService())
const cardBackgroundUploadService = shallowReactive(new CardBackgroundUploadService())
const projectService = new ProjectService()
const configStore = useConfigStore()

// Card images reuse the same provider config as the full background (see
// ProjectSettingsDropdown.vue's backgroundsEnabled computed for the same choice).
const unsplashCardImageEnabled = computed(() => configStore.enabledBackgroundProviders.includes('unsplash'))
const uploadCardImageEnabled = computed(() => configStore.enabledBackgroundProviders.includes('upload'))

function hasCardImage(project: IProject) {
	return !!project.cardBackgroundInformation
}

function removeCardImage(project: IProject) {
	return projectService.removeCardBackground(project)
}
</script>
