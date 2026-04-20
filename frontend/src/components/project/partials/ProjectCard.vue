<template>
	<div
		class="project-card"
		:class="{
			'has-light-text': background !== null,
			'has-background': blurHashUrl !== '' || background !== null
		}"
		:style="{
			'border-inline-start': project.hexColor ? `0.25rem solid ${project.hexColor}` : undefined,
			'background-image': blurHashUrl !== '' ? `url(${blurHashUrl})` : undefined,
		}"
	>
		<div
			class="project-background background-fade-in"
			:class="{'is-visible': background}"
			:style="{'background-image': background !== null ? `url(${background})` : undefined}"
		/>
		<span
			v-if="project.isArchived"
			class="is-archived"
		>{{ $t('project.archived') }}</span>

		<div
			class="project-title"
			aria-hidden="true"
		>
			<span
				v-if="project.id < -1"
				class="saved-filter-icon icon"
			>
				<Icon icon="filter" />
			</span>
			{{ getProjectTitle(project) }}
		</div>
		<BaseButton
			class="project-button"
			:aria-label="project.title"
			:title="textOnlyDescription"
			:to="{
				name: 'project.index',
				params: { projectId: project.id}
			}"
		/>
		<BaseButton
			v-if="!project.isArchived && project.id > -1"
			class="favorite"
			:class="{'is-favorite': project.isFavorite}"
			@click.prevent.stop="projectStore.toggleProjectFavorite(project)"
		>
			<Icon :icon="project.isFavorite ? 'star' : ['far', 'star']" />
		</BaseButton>
	</div>
</template>

<script lang="ts" setup>
import {computed} from 'vue'
import type {IProject} from '@/modelTypes/IProject'

import BaseButton from '@/components/base/BaseButton.vue'

import {useProjectBackground} from '@/composables/useProjectBackground'
import {useProjectStore} from '@/stores/projects'
import {getProjectTitle} from '@/helpers/getProjectTitle'

const props = defineProps<{
	project: IProject,
}>()

const {background, blurHashUrl} = useProjectBackground(() => props.project)

const projectStore = useProjectStore()

const textOnlyDescription = computed(() => {
	return props.project.description ? props.project.description.replace(/<[^>]*>/g, '') : ''
})
</script>

<style lang="scss" scoped>
.project-card {
	--project-card-padding: 1rem;
	background: var(--vk-bg-panel);
	padding: var(--project-card-padding);
	border-radius: 12px;
	border: .5px solid var(--vk-border-mid);
	box-shadow: none;
	transition: border-color .2s ease, transform .2s ease, background-color .2s ease;
	position: relative;
	overflow: hidden; // hide background

	display: flex;
	justify-content: space-between;
	flex-wrap: wrap;

	&::before {
		content: '';
		position: absolute;
		inset-block-start: 0;
		inset-inline-start: 0;
		inset-inline-end: 0;
		block-size: 2px;
		background: linear-gradient(90deg, #6c63f5 0%, #34d399 50%, #f59e0b 100%);
		opacity: .5;
	}

	&:hover {
		border-color: var(--vk-border-mid);
		transform: translateY(-2px);
		background: var(--vk-bg-hover);
	}

	&:active,
	&:focus {
		transform: translateY(0);
		border-color: var(--vk-accent);
	}

	> * {
		// so the elements are on top of the background
		position: relative;
	}
}

.has-background,
.project-background {
	background-size: cover;
	background-repeat: no-repeat;
	background-position: center;
}

.project-background,
.project-button {
	position: absolute;
	inset-block-start: 0;
	inset-inline-end: 0;
	inset-block-end: 0;
	inset-inline-start: 0;
}

.is-archived {
	font-size: .7rem;
	float: inline-start;
	padding: 2px 7px;
	border-radius: 999px;
	background: var(--vk-bg-panel);
	border: .5px solid var(--vk-border-mid);
	color: var(--vk-text-secondary);
	font-weight: 600;
	letter-spacing: .02em;
}

.project-title {
	align-self: flex-end;
	font-family: $vikunja-font;
	font-weight: 500;
	font-size: 1.125rem;
	line-height: var(--title-line-height);
	color: var(--vk-text-primary);
	inline-size: 100%;
	margin-block-end: 0;
	max-block-size: calc(100% - (var(--project-card-padding) + 1rem)); // padding & height of the "is archived" badge
	overflow: hidden;
	text-overflow: ellipsis;
	word-break: break-word;

	display: -webkit-box;
	-webkit-line-clamp: 3;
	-webkit-box-orient: vertical;
	text-wrap: balance;
}

.has-light-text .project-title {
	color: #f3f4f6;
}

.has-background .project-title {
	text-shadow:
		0 8px 20px rgb(0 0 0 / 70%),
		0 2px 6px rgb(0 0 0 / 60%);
	color: var(--white);
}

.favorite {
	position: absolute;
	inset-block-start: var(--project-card-padding);
	inset-inline-end: var(--project-card-padding);
	transition: opacity $transition, color $transition, border-color $transition, background-color $transition;
	opacity: 1;
	background: var(--vk-bg-panel);
	border: .5px solid var(--vk-border-mid);
	border-radius: 999px;
	inline-size: 1.9rem;
	block-size: 1.9rem;
	display: inline-flex;
	align-items: center;
	justify-content: center;
	color: var(--vk-text-secondary);

	&:hover {
		border-color: #f59e0b;
		background: rgb(245 158 11 / 12%);
		color: #fbbf24;
	}

	&.is-favorite {
		display: inline-block;
		opacity: 1;
		border-color: #f59e0b;
		background: rgb(245 158 11 / 12%);
		color: #fbbf24;
	}
}

@media(hover: hover) and (pointer: fine) {
	.project-card .favorite {
		opacity: 0;
	}

	.project-card:hover .favorite {
		opacity: 1;
	}
}

.background-fade-in {
  opacity: 0;
  transition: opacity $transition;
  transition-delay: $transition-duration * 2; // To fake an appearing background

  &.is-visible {
    opacity: 1;
  }
}

.saved-filter-icon {
	color: var(--vk-text-secondary);
	font-size: .75em;
}
</style>
