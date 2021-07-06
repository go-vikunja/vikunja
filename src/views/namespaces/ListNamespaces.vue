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
				<list-card
					v-for="l in n.lists"
					:key="`l${l.id}`"
					:list="l"
					:show-archived="showArchived"
				/>
			</div>
		</div>
	</div>
</template>

<script>
import {mapState} from 'vuex'
import Fancycheckbox from '../../components/input/fancycheckbox'
import {LOADING} from '@/store/mutation-types'
import ListCard from '@/components/list/partials/list-card'

export default {
	name: 'ListNamespaces',
	components: {
		ListCard,
		Fancycheckbox,
	},
	data() {
		return {
			showArchived: false,
		}
	},
	created() {
		this.showArchived = JSON.parse(localStorage.getItem('showArchived')) ?? false
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
		saveShowArchivedState() {
			localStorage.setItem('showArchived', JSON.stringify(this.showArchived))
		},
	},
}
</script>
