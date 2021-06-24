<template>
	<div class="content namespaces-list loader-container" :class="{'is-loading': loading}">
		<x-button :to="{name: 'namespace.create'}" class="new-namespace" icon="plus">
			{{ $t('namespace.create.title') }}
		</x-button>
		<x-button :to="{name: 'filters.create'}" class="new-namespace" icon="filter">
			{{ $t('filters.create.title') }}
		</x-button>

		<fancycheckbox class="show-archived-check" v-model="showArchived" @change="saveShowArchivedState">
			{{ $t('namespace.showArchived') }}
		</fancycheckbox>

		<p class="has-text-centered has-text-grey mt-4 is-italic" v-if="namespaces.length === 0">
			{{ $t('namespace.noneAvailable') }}
			<router-link :to="{name: 'namespace.create'}">
				{{ $t('namespace.create.title') }}.
			</router-link>
		</p>

		<div :key="`n${n.id}`" class="namespace" v-for="n in namespaces">
			<x-button
				:to="{name: 'list.create', params: {id:  n.id}}"
				class="is-pulled-right"
				type="secondary"
				v-if="n.id > 0 && n.lists.length > 0"
				icon="plus"
			>
				{{ $t('list.create.header') }}
			</x-button>
			<x-button
				:to="{name: 'namespace.settings.archive', params: {id:  n.id}}"
				class="is-pulled-right mr-4"
				type="secondary"
				v-if="n.isArchived"
				icon="archive"
			>
				{{ $t('namespace.unarchive') }}
			</x-button>

			<h1>
				<span>{{ n.title }}</span>
				<span class="is-archived" v-if="n.isArchived">
					{{ $t('namespace.archived') }}
				</span>
			</h1>

			<p class="has-text-centered has-text-grey mt-4 is-italic" v-if="n.lists.length === 0">
				{{ $t('namespace.noLists') }}
				<router-link :to="{name: 'list.create', params: {id:  n.id}}">
					{{ $t('namespace.createList') }}
				</router-link>
			</p>

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
								{{ $t('namespace.archived') }}
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
		this.showArchived = JSON.parse(localStorage.getItem('showArchived')) ?? false
		this.loadBackgroundsForLists()
	},
	mounted() {
		this.setTitle(this.$t('namespace.title'))
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
								this.error(e)
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
				.catch(e => this.error(e))
		},
		saveShowArchivedState() {
			localStorage.setItem('showArchived', JSON.stringify(this.showArchived))
		},
	},
}
</script>
