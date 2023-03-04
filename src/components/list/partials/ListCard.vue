<template>
	<div
		class="list-card"
		:class="{
			'has-light-text': background !== null,
			'has-background': blurHashUrl !== '' || background !== null
		}"
		:style="{
			'border-left': list.hexColor ? `0.25rem solid ${list.hexColor}` : undefined,
			'background-image': blurHashUrl !== '' ? `url(${blurHashUrl})` : undefined,
		}"
	>
		<div
			class="list-background background-fade-in"
			:class="{'is-visible': background}"
			:style="{'background-image': background !== null ? `url(${background})` : undefined}"
		/>
		<span v-if="list.isArchived" class="is-archived" >{{ $t('namespace.archived') }}</span>

		<div class="list-title" aria-hidden="true">{{ list.title }}</div>
		<BaseButton
			class="list-button"
			:aria-label="list.title"
			:title="list.description"
			:to="{
				name: 'list.index',
				params: { listId: list.id}
			}"
		/>
		<BaseButton
			v-if="!list.isArchived"
			class="favorite"
			:class="{'is-favorite': list.isFavorite}"
			@click.prevent.stop="listStore.toggleListFavorite(list)"
		>
			<icon :icon="list.isFavorite ? 'star' : ['far', 'star']" />
		</BaseButton>
	</div>
</template>

<script lang="ts" setup>
import {toRef, type PropType} from 'vue'

import type {IList} from '@/modelTypes/IList'

import BaseButton from '@/components/base/BaseButton.vue'

import {useListBackground} from './useListBackground'
import {useListStore} from '@/stores/lists'

const props = defineProps({
	list: {
		type: Object as PropType<IList>,
		required: true,
	},
})

const {background, blurHashUrl} = useListBackground(toRef(props, 'list'))

const listStore = useListStore()
</script>

<style lang="scss" scoped>
.list-card {
	--list-card-padding: 1rem;
	background: var(--white);
	padding: var(--list-card-padding);
	border-radius: $radius;
	box-shadow: var(--shadow-sm);
	transition: box-shadow $transition;
	position: relative;
	overflow: hidden; // hide background

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
.list-background {
	background-size: cover;
	background-repeat: no-repeat;
	background-position: center;
}

.list-background,
.list-button {
	position: absolute;
	top: 0;
	right: 0;
	bottom: 0;
	left: 0;
}

.is-archived {
	font-size: .75rem;
	float: left;
}

.list-title {
	align-self: flex-end;
	font-family: $vikunja-font;
	font-weight: 400;
	font-size: 1.5rem;
	line-height: var(--title-line-height);
	color: var(--text);
	width: 100%;
	margin-bottom: 0;
	max-height: calc(100% - (var(--list-card-padding) + 1rem)); // padding & height of the "is archived" badge
	overflow: hidden;
	text-overflow: ellipsis;
	word-break: break-word;

	display: -webkit-box;
	-webkit-line-clamp: 3;
	-webkit-box-orient: vertical;
}

.has-light-text .list-title {
	color: var(--grey-100);
}

.has-background .list-title {
	text-shadow:
		0 0 10px var(--black),
		1px 1px 5px var(--grey-700),
		-1px -1px 5px var(--grey-700);
	color: var(--white);
}

.favorite {
	position: absolute;
	top: var(--list-card-padding);
	right: var(--list-card-padding);
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
	.list-card .favorite {
		opacity: 0;
	}

	.list-card:hover .favorite {
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
</style>
