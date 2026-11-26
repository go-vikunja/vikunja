<template>
	<div
		class="project-card"
		:class="{
			'has-background': Boolean(blurHashUrl || background),
			'is-light-background': backgroundIsLight,
		}"
		:style="{
			'border-inline-start': project.hex_color ? `0.25rem solid ${project.hex_color}` : undefined,
			'background-image': blurHashUrl ? `url(${blurHashUrl})` : undefined,
		}"
	>
		<div
			class="project-background background-fade-in"
			:class="{'is-visible': background}"
			:style="{'background-image': background ? `url(${background})` : undefined}"
		/>
		<span
			v-if="project.is_archived"
			class="is-archived"
		>{{ $t('project.archived') }}</span>

		<div
			class="project-title"
			aria-hidden="true"
		>
			<span class="project-title-text">
				<span
					v-if="project.id < -1"
					class="saved-filter-icon icon"
				>
					<Icon icon="filter" />
				</span>
				{{ titleParts.head }}<span :class="{'title-tail': project.is_shared}">{{ titleParts.tail }}<span
					v-if="project.is_shared"
					v-tooltip="$t('project.sharedWithOthers')"
					class="shared-icon"
				><Icon icon="users" /></span></span>
			</span>
		</div>
		<BaseButton
			class="project-button"
			:aria-label="buttonLabel"
			:title="textOnlyDescription"
			:to="{
				name: 'project.index',
				params: { projectId: project.id}
			}"
		/>
		<div class="project-actions">
			<BaseButton
				v-if="!project.is_archived && project.id > -1"
				class="favorite"
				:aria-label="project.is_favorite ? $t('project.unfavorite') : $t('project.favorite')"
				:class="{'is-favorite': project.is_favorite}"
				@click.prevent.stop="toggleProjectFavorite"
			>
				<Icon :icon="project.is_favorite ? 'star' : ['far', 'star']" />
			</BaseButton>
			<ProjectSettingsDropdown
				v-if="typeof project.max_permission === 'number' && project.max_permission > PERMISSIONS.READ"
				class="settings-dropdown"
				:project="project"
			/>
		</div>
	</div>
</template>

<script lang="ts" setup>
import {computed} from 'vue'
import {useI18n} from 'vue-i18n'
import {usePatchProjectFavoriteMutation, type ProjectResponse} from '@/client/queries/projects'

import BaseButton from '@/components/base/BaseButton.vue'
import ProjectSettingsDropdown from '@/components/project/ProjectSettingsDropdown.vue'
import {PERMISSIONS} from '@/constants/permissions'

import {useProjectBackground} from '@/composables/useProjectBackground'
import {getProjectTitle} from '@/helpers/getProjectTitle'
import {htmlToPreviewText} from '@/helpers/htmlToPreviewText'
import {blurHashIsLight} from '@/helpers/color/blurHashIsLight'

const props = defineProps<{
	project: ProjectResponse,
}>()

const {background, blurHashUrl} = useProjectBackground(() => props.project)
const backgroundIsLight = computed(() => blurHashIsLight(props.project.background_blur_hash ?? ''))

const favoriteMutation = usePatchProjectFavoriteMutation()

async function toggleProjectFavorite() {
	if (props.project.id <= 0 || props.project.is_archived) {
		return
	}
	await favoriteMutation.mutateAsync({id: props.project.id, isFavorite: !props.project.is_favorite})
}

const {t} = useI18n()
// The title is aria-hidden, so the shared state has to be in the link label.
const buttonLabel = computed(() => props.project.is_shared
	? `${props.project.title}, ${t('project.sharedWithOthers')}`
	: props.project.title)

// Keeps the icon on the same line as the last word.
const titleParts = computed(() => {
	const title = getProjectTitle(props.project)
	const split = title.lastIndexOf(' ') + 1
	return {head: title.slice(0, split), tail: title.slice(split)}
})

const textOnlyDescription = computed(() => htmlToPreviewText(props.project.description ?? ''))
</script>

<style lang="scss" scoped>
.project-card {
	--project-card-padding: 1rem;
	background: var(--white);
	padding: var(--project-card-padding);
	border-radius: $radius;
	box-shadow: var(--shadow-sm);
	transition: box-shadow $transition;
	position: relative;

	display: flex;
	justify-content: space-between;
	flex-wrap: wrap;

	&:hover {
		box-shadow: var(--shadow-md);
	}

	&:active,
	&:focus {
		box-shadow: var(--shadow-xs) !important;
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

// Clip here instead of on the card so the settings dropdown can overflow it.
.project-background {
	border-radius: inherit;
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
	font-size: .75rem;
	float: inline-start;
}

.project-title {
	align-self: flex-end;
	font-family: $vikunja-font;
	font-weight: 400;
	font-size: 1.5rem;
	line-height: var(--title-line-height);
	color: var(--text);
	inline-size: 100%;
	margin-block-end: 0;
	max-block-size: calc(100% - (var(--project-card-padding) + 1rem)); // padding & height of the "is archived" badge
	// The shadow filter stacks the title under .project-button; lift it so the shared icon gets hover.
	z-index: 1;
	pointer-events: none;
}

.project-title-text {
	overflow: hidden;
	text-overflow: ellipsis;
	word-break: break-word;
	// Room for descenders; more would reveal the clamped 4th line.
	padding-block-end: .25rem;
	margin-block-end: -.25rem;

	display: -webkit-box;
	-webkit-line-clamp: 3;
	-webkit-box-orient: vertical;
}

// Fixed colours, not theme variables: the image looks the same in both themes.
.has-background {
	--on-image-text: hsl(0 0% 100%);
	// Chained drop-shadows compound, hence the light values.
	--on-image-shadow:
		drop-shadow(0 0 2px hsl(216.9 19.1% 26.7%))
		drop-shadow(0 0 6px rgb(0 0 0 / .6));

	&.is-light-background {
		--on-image-text: hsl(220.9 39.3% 11%);
		--on-image-shadow:
			drop-shadow(0 0 2px hsl(0 0% 100%))
			drop-shadow(0 0 6px rgb(255 255 255 / .7));
	}
}

// Not text-shadow: the clamp's overflow would clip it.
.has-background .project-title,
.has-background .is-archived,
.has-background .favorite,
.has-background :deep(.dropdown-trigger) {
	filter: var(--on-image-shadow);
	color: var(--on-image-text);
}

.has-background .shared-icon {
	opacity: .8;
}

.project-actions {
	position: absolute;
	inset-block-start: var(--project-card-padding);
	inset-inline-end: var(--project-card-padding);
	display: flex;
	align-items: center;
	gap: .5rem;
}

.favorite,
.settings-dropdown {
	transition: opacity $transition, color $transition;
	opacity: 1;

	&:hover {
		color: var(--warning);
	}

	&.is-favorite {
		display: inline-block;
		opacity: 1;
		color: var(--warning);
	}
}

@media(hover: hover) and (pointer: fine) {
	.project-card .favorite,
	.project-card .settings-dropdown {
		opacity: 0;
	}

	.project-card:hover .favorite,
	.project-card:hover .settings-dropdown,
	.project-card:focus-within .settings-dropdown,
	.settings-dropdown:has([aria-expanded="true"]) {
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

.title-tail {
	white-space: nowrap;
}

.shared-icon {
	pointer-events: auto;
	margin-inline-start: .375rem;
	font-size: .875rem;
	opacity: .5;
}

.saved-filter-icon {
	color: var(--grey-300);
	font-size: .75em;
}
</style>
