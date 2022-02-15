<template>
	<div class="content loader-container" :class="{'is-loading': loading}" v-cy="'namespaces-list'">
		<header class="namespace-header">
			<fancycheckbox v-model="showArchived" @change="saveShowArchivedState" v-cy="'show-archived-check'">
				{{ $t('namespace.showArchived') }}
			</fancycheckbox>

			<div class="action-buttons">
				<x-button :to="{name: 'filters.create'}" icon="filter">
					{{ $t('filters.create.title') }}
				</x-button>
				<x-button :to="{name: 'namespace.create'}" icon="plus" v-cy="'new-namespace'">
					{{ $t('namespace.create.title') }}
				</x-button>
			</div>
		</header>

		<p class="has-text-centered has-text-grey mt-4 is-italic" v-if="namespaces.length === 0">
			{{ $t('namespace.noneAvailable') }}
			<router-link :to="{name: 'namespace.create'}">
				{{ $t('namespace.create.title') }}.
			</router-link>
		</p>

		<section :key="`n${n.id}`" class="namespace" v-for="n in namespaces">
			<x-button
				:to="{name: 'list.create', params: {namespaceId:  n.id}}"
				class="is-pulled-right"
				variant="secondary"
				v-if="n.id > 0 && n.lists.length > 0"
				icon="plus"
			>
				{{ $t('list.create.header') }}
			</x-button>
			<x-button
				:to="{name: 'namespace.settings.archive', params: {id:  n.id}}"
				class="is-pulled-right mr-4"
				variant="secondary"
				v-if="n.isArchived"
				icon="archive"
			>
				{{ $t('namespace.unarchive') }}
			</x-button>

			<h2 class="namespace-title">
				<span v-cy="'namespace-title'">{{ getNamespaceTitle(n) }}</span>
				<span class="is-archived" v-if="n.isArchived">
					{{ $t('namespace.archived') }}
				</span>
			</h2>

			<p class="has-text-centered has-text-grey mt-4 is-italic" v-if="n.lists.length === 0">
				{{ $t('namespace.noLists') }}
				<router-link :to="{name: 'list.create', params: {namespaceId:  n.id}}">
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
		</section>
	</div>
</template>

<script lang="ts">
import {mapState} from 'vuex'
import Fancycheckbox from '../../components/input/fancycheckbox.vue'
import {LOADING} from '@/store/mutation-types'
import ListCard from '@/components/list/partials/list-card.vue'

export default {
	name: 'ListNamespaces',
	components: {
		ListCard,
		Fancycheckbox,
	},
	data() {
		return {
			showArchived: JSON.parse(localStorage.getItem('showArchived')) ?? false,
		}
	},
	mounted() {
		this.setTitle(this.$t('namespace.title'))
	},
	computed: mapState({
		namespaces(state) {
			return state.namespaces.namespaces.filter(n => this.showArchived ? true : !n.isArchived)
			// return state.namespaces.namespaces.filter(n => this.showArchived ? true : !n.isArchived).map(n => {
			// 	n.lists = n.lists.filter(l => !l.isArchived)
			// 	return n
			// })
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

<style lang="scss" scoped>
.namespace-header {
	display: flex;
	justify-content: space-between;
	align-items: center;
	gap: 1rem;

	@media screen and (max-width: $tablet) {
		flex-direction: column;
	}
}

.action-buttons {
	display: flex;
	justify-content: space-between;
	gap: 1rem;

	@media screen and (max-width: $tablet) {
		width: 100%;
		flex-direction: column;
		align-items: stretch;
	}
}

.namespace {
	& + & {
		margin-top: 1rem;
	}
}

.namespace-title {
	display: flex;
	align-items: center;
}

.is-archived {
	font-size: 0.75rem;
	border: 1px solid var(--grey-500);
	color: $grey !important;
	padding: 2px 4px;
	border-radius: 3px;
	font-family: $vikunja-font;
	background: var(--white-translucent);
	margin-left: .5rem;
}

.lists {
	display: flex;
	flex-flow: row wrap;
}
</style>