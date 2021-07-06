<template>
	<router-link
		:class="{
			'has-light-text': !colorIsDark(list.hexColor),
			'has-background': backgroundResolver(list.id) !== null
		}"
		:style="{
			'background-color': list.hexColor,
			'background-image': backgroundResolver(list.id) !== null ? `url(${backgroundResolver(list.id)})` : false,
		}"
		:to="{ name: 'list.index', params: { listId: list.id} }"
		class="list-card"
		tag="span"
		v-if="showArchived ? true : !list.isArchived"
	>
		<div class="is-archived-container">
			<span class="is-archived" v-if="list.isArchived">
				{{ $t('namespace.archived') }}
			</span>
			<span
				:class="{'is-favorite': list.isFavorite, 'is-archived': list.isArchived}"
				@click.stop="toggleFavoriteList(list)"
				class="favorite">
				<icon icon="star" v-if="list.isFavorite"/>
				<icon :icon="['far', 'star']" v-else/>
			</span>
		</div>
		<div class="title">{{ list.title }}</div>
	</router-link>
</template>

<script>
export default {
	name: 'list-card',
	props: {
		list: {
			required: true,
		},
		showArchived: {
			default: false,
			type: Boolean,
		},
		// A function, returning a background blob or null if none exists for that list.
		// Receives the list id as parameter.
		backgroundResolver: {
			required: true,
		},
	},
	methods: {
		toggleFavoriteList(list) {
			// The favorites pseudo list is always favorite
			// Archived lists cannot be marked favorite
			if (list.id === -1 || list.isArchived) {
				return
			}
			this.$store.dispatch('lists/toggleListFavorite', list)
				.catch(e => this.error(e))
		},
	},
}
</script>
