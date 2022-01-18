<template>
	<router-link
		:class="{
			'has-light-text': !colorIsDark(list.hexColor),
			'has-background': background !== null
		}"
		:style="{
			'background-color': list.hexColor,
			'background-image': background !== null ? `url(${background})` : false,
		}"
		:to="{ name: 'list.index', params: { listId: list.id} }"
		class="list-card"
		v-if="list !== null && (showArchived ? true : !list.isArchived)"
	>
		<div class="is-archived-container">
			<span class="is-archived" v-if="list.isArchived">
				{{ $t('namespace.archived') }}
			</span>
			<span
				:class="{'is-favorite': list.isFavorite, 'is-archived': list.isArchived}"
				@click.stop="toggleFavoriteList(list)"
				class="favorite">
				<icon :icon="list.isFavorite ? 'star' : ['far', 'star']" />
			</span>
		</div>
		<div class="title">{{ list.title }}</div>
	</router-link>
</template>

<script lang="ts" setup>
import {PropType, ref, watch} from 'vue'
import {useStore} from 'vuex'

import ListService from '@/services/list'

import {colorIsDark} from '@/helpers/color/colorIsDark'
import ListModel from '@/models/list'

const background = ref<string | null>(null)
const backgroundLoading = ref(false)

const props = defineProps({
	list: {
		type: Object as PropType<ListModel>,
		required: true,
	},
	showArchived: {
		default: false,
		type: Boolean,
	},
})

watch(props.list, loadBackground, { immediate: true })

async function loadBackground() {
	if (props.list === null || !props.list.backgroundInformation || backgroundLoading.value) {
		return
	}

	backgroundLoading.value = true

	const listService = new ListService()
	try {
		background.value = await listService.background(props.list)
	} finally {
		backgroundLoading.value = false
	}
}

const store = useStore()

function toggleFavoriteList(list: ListModel) {
	// The favorites pseudo list is always favorite
	// Archived lists cannot be marked favorite
	if (list.id === -1 || list.isArchived) {
		return
	}
	store.dispatch('lists/toggleListFavorite', list)
}
</script>

<style lang="scss" scoped>
.list-card {
  cursor: pointer;
  width: calc((100% - #{($lists-per-row - 1) * 1rem}) / #{$lists-per-row});
  height: $list-height;
  background: var(--white);
  margin: 0 $list-spacing $list-spacing 0;
  padding: 1rem;
  border-radius: $radius;
  box-shadow: var(--shadow-sm);
  transition: box-shadow $transition;

  display: flex;
  justify-content: space-between;
  flex-wrap: wrap;

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

  .is-archived-container {
    width: 100%;
    text-align: right;

    .is-archived {
      font-size: .75rem;
      float: left;
    }
  }

  .title {
    align-self: flex-end;
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

  &.has-light-text .title {
    color: var(--light);
  }

  &.has-background {
    background-size: cover;
    background-repeat: no-repeat;
    background-position: center;

    .title {
      text-shadow: 0 0 10px var(--black), 1px 1px 5px var(--grey-700), -1px -1px 5px var(--grey-700);
      color: var(--white);
    }
  }

  .favorite {
    transition: opacity $transition, color $transition;
    opacity: 0;

    &:hover {
      color: var(--warning);
    }

    &.is-archived {
      display: none;
    }

    &.is-favorite {
      display: inline-block;
      opacity: 1;
      color: var(--warning);
    }
  }

  &:hover .favorite {
    opacity: 1;
  }
}
</style>