<template>
	<div class="content namespaces-list loader-container" :class="{'is-loading': loading}">
		<x-button :to="{name: 'namespace.create'}" class="new-namespace" icon="plus">
			Create namespace
		</x-button>
		<x-button :to="{name: 'filters.create'}" class="new-namespace" icon="filter">
			Create saved filter
		</x-button>

		<fancycheckbox class="show-archived-check" v-model="showArchived">
			Show Archived
		</fancycheckbox>

		<div :key="`n${n.id}`" class="namespace" v-for="n in namespaces">
			<x-button
				:to="{name: 'list.create', params: {id:  n.id}}"
				class="is-pulled-right"
				type="secondary"
				v-if="n.id > 0"
				icon="plus"
			>
				Create list
			</x-button>

			<h1>
				<span>{{ n.title }}</span>
				<span class="is-archived" v-if="n.isArchived">
					Archived
				</span>
			</h1>

			<div class="lists">
				<template v-for="l in n.lists">
					<router-link
						:class="{
							'has-light-text': !colorIsDark(l.hexColor),
							'has-background': typeof backgrounds[l.id] !== 'undefined',
						}"
						:key="`l${l.id}`"
						:style="{
							'background-color': l.hexColor,
							'background-image': typeof backgrounds[l.id] !== 'undefined' ? `url(${backgrounds[l.id]})` : false,
						}"
						:to="{ name: 'list.index', params: { listId: l.id} }"
						class="list"
						tag="span"
						v-if="showArchived ? true : !l.isArchived"
					>
						<div class="is-archived-container">
							<span class="is-archived" v-if="l.isArchived">
								Archived
							</span>
							<span
								:class="{'is-favorite': l.isFavorite, 'is-archived': l.isArchived}"
								@click.stop="toggleFavoriteList(l)"
								class="favorite">
								<icon icon="star" v-if="l.isFavorite"/>
								<icon :icon="['far', 'star']" v-else/>
							</span>
						</div>
						<div class="title">{{ l.title }}</div>
					</router-link>
				</template>
			</div>
		</div>
	</div>
</template>

<script>
import {mapState} from 'vuex'
import ListService from '../../services/list'
import Fancycheckbox from '../../components/input/fancycheckbox'
import {LOADING} from '@/store/mutation-types'

export default {
	name: 'ListNamespaces',
	components: {
		Fancycheckbox,
	},
	data() {
		return {
			showArchived: false,
			// listId is the key, the object is the background blob
			backgrounds: {},
		}
	},
	created() {
		this.loadBackgroundsForLists()
	},
	mounted() {
		this.setTitle('Namespaces & Lists')
	},
	computed: mapState({
		namespaces(state) {
			return state.namespaces.namespaces.filter(n => this.showArchived ? true : !n.isArchived)
		},
		loading: LOADING,
	}),
	methods: {
		loadBackgroundsForLists() {
			const listService = new ListService()
			this.namespaces.forEach(n => {
				n.lists.forEach(l => {
					if (l.backgroundInformation) {
						listService.background(l)
							.then(b => {
								this.$set(this.backgrounds, l.id, b)
							})
							.catch(e => {
								this.error(e, this)
							})
					}
				})
			})
		},
		toggleFavoriteList(list) {
			// The favorites pseudo list is always favorite
			// Archived lists cannot be marked favorite
			if (list.id === -1 || list.isArchived) {
				return
			}
			this.$store.dispatch('lists/toggleListFavorite', list)
				.catch(e => this.error(e, this))
		},
	},
}
</script>
