<template>
	<router-link
		:class="{
			'has-light-text': !colorIsDark(list.hexColor) || background !== null,
			'has-background': blurHashUrl !== '' || background !== null,
		}"
		:style="{
			'border-color': `${list.hexColor}`,
			'background-image': blurHashUrl !== null ? `url(${blurHashUrl})` : false,
		}"
		:to="{ name: 'list.index', params: { listId: list.id} }"
		class="list-card"
		v-if="list !== null && (showArchived ? true : !list.isArchived)"
	>
		<div
			class="list-background background-fade-in"
			:class="{'is-visible': background}"
			:style="{'background-image': background !== null ? `url(${background})` : undefined}"
		/>
		<div class="list-content">
			<span class="is-archived" v-if="list.isArchived">
				{{ $t('namespace.archived') }}
			</span>
			<BaseButton
				v-else
				:class="{'is-favorite': list.isFavorite}"
				@click.stop="listStore.toggleListFavorite(list)"
				class="favorite"
			>
				<icon :icon="list.isFavorite ? 'star' : ['far', 'star']"/>
			</BaseButton>

			<div class="title">{{ list.title }}</div>
		</div>
	</router-link>
</template>

<script lang="ts" setup>
import {type PropType, ref, watch} from 'vue'

import ListService from '@/services/list'
import {getBlobFromBlurHash} from '@/helpers/getBlobFromBlurHash'

import {colorIsDark} from '@/helpers/color/colorIsDark'

import BaseButton from '@/components/base/BaseButton.vue'
import type {IList} from '@/modelTypes/IList'
import {useListStore} from '@/stores/lists'

const background = ref<string | null>(null)
const backgroundLoading = ref(false)
const blurHashUrl = ref('')

const props = defineProps({
	list: {
		type: Object as PropType<IList>,
		required: true,
	},
	showArchived: {
		default: false,
		type: Boolean,
	},
})

watch(props.list, loadBackground, {immediate: true})

async function loadBackground() {
	if (props.list === null || !props.list.backgroundInformation || backgroundLoading.value) {
		return
	}

	const blurHash = await getBlobFromBlurHash(props.list.backgroundBlurHash)
	if (blurHash) {
		blurHashUrl.value = window.URL.createObjectURL(blurHash)
	}

	backgroundLoading.value = true

	const listService = new ListService()
	try {
		background.value = await listService.background(props.list)
	} finally {
		backgroundLoading.value = false
	}
}

const listStore = useListStore()

</script>

<style lang="scss" scoped>
.list-card {
	cursor: pointer;
	width: calc((100% - #{($lists-per-row - 1) * 1rem}) / #{$lists-per-row});
	height: $list-height;
	border-left-width: 0.8rem;
	border-left-style: solid;
	background: var(--white);
	margin: 0 $list-spacing $list-spacing 0;
	border-radius: $radius;
	box-shadow: var(--shadow-sm);
	transition: box-shadow $transition;
	position: relative;
	overflow: hidden;

	&.has-light-text .title {
		color: var(--grey-100) !important;
	}

	&.has-background,
	.list-background {
		background-size: cover;
		background-repeat: no-repeat;
		background-position: center;
	}

	&.has-background .title {
		text-shadow: 0 0 10px var(--black), 1px 1px 5px var(--grey-700), -1px -1px 5px var(--grey-700);
		color: var(--white);
	}

	.list-background {
		position: absolute;
		top: 0;
		right: 0;
		bottom: 0;
		left: 0;
	}

	&:hover {
		box-shadow: var(--shadow-md);
	}

	&:active,
	&:focus,
	&:focus:not(:active) {
		box-shadow: var(--shadow-xs) !important;
	}

	@media screen and (min-width: $widescreen) {
		&:nth-child(#{$lists-per-row}n) {
			margin-right: 0;
		}
	}

	@media screen and (max-width: $widescreen) and (min-width: $tablet) {
		$lists-per-row: 3;
		& {
			width: calc((100% - #{($lists-per-row - 1) * 1rem}) / #{$lists-per-row});
		}

		&:nth-child(#{$lists-per-row}n) {
			margin-right: 0;
		}
	}

	@media screen and (max-width: $tablet) {
		$lists-per-row: 2;
		& {
			width: calc((100% - #{($lists-per-row - 1) * 1rem}) / #{$lists-per-row});
		}

		&:nth-child(#{$lists-per-row}n) {
			margin-right: 0;
		}
	}

	@media screen and (max-width: $mobile) {
		$lists-per-row: 1;
		& {
			width: 100%;
			margin-right: 0;
		}
	}

	.list-content {
		display: flex;
		align-content: flex-start;
		flex-wrap: wrap;
		row-gap: 0.8rem;
		padding: 1rem;
		position: absolute;
		height: 100%;
		width: 100%;


		.is-archived {
			font-size: .75rem;
		}

		.favorite {
			margin-left: auto;
			transition: opacity $transition, color $transition;
			opacity: 0;
			display: block;

			&:hover,
			&.is-favorite {
				color: var(--warning);
			}
		}

		.favorite.is-favorite,
		&:hover .favorite {
			opacity: 1;
		}

		.title {
			align-self: flex-start;
			font-family: $vikunja-font;
			font-weight: 400;
			font-size: 1.5rem;
			color: var(--text);
			width: 100%;
			margin-bottom: 0;
			max-height: calc(100% - 2rem); // 1rem padding, 1rem height of the "is archived" badge
			overflow: hidden;
			text-overflow: ellipsis;

			display: -webkit-box;
			-webkit-line-clamp: 3;
			-webkit-box-orient: vertical;
		}
	}
}
</style>