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
		tag="span"
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
				<icon icon="star" v-if="list.isFavorite"/>
				<icon :icon="['far', 'star']" v-else/>
			</span>
		</div>
		<div class="title">{{ list.title }}</div>
	</router-link>
</template>

<script>
import ListService from '@/services/list'

export default {
	name: 'list-card',
	data() {
		return {
			background: null,
			backgroundLoading: false,
		}
	},
	props: {
		list: {
			required: true,
		},
		showArchived: {
			default: false,
			type: Boolean,
		},
	},
	watch: {
		list() {
			this.loadBackground()
		},
	},
	created() {
		this.loadBackground()
	},
	methods: {
		loadBackground() {
			if (this.list === null || !this.list.backgroundInformation || this.backgroundLoading) {
				return
			}

			this.backgroundLoading = true

			const listService = new ListService()
			listService.background(this.list)
				.then(b => {
					this.$set(this, 'background', b)
				})
				.catch(e => {
					this.$message.error(e)
				})
				.finally(() => this.backgroundLoading = false)
		},
		toggleFavoriteList(list) {
			// The favorites pseudo list is always favorite
			// Archived lists cannot be marked favorite
			if (list.id === -1 || list.isArchived) {
				return
			}
			this.$store.dispatch('lists/toggleListFavorite', list)
				.catch(e => this.$message.error(e))
		},
	},
}
</script>
