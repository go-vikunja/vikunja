<template>
	<div class="content namespaces-list">
		<router-link :to="{name: 'namespace.create'}" class="button is-success new-namespace">
			<span class="icon is-small">
				<icon icon="plus"/>
			</span>
			Create new namespace
		</router-link>

		<fancycheckbox v-model="showArchived" class="show-archived-check">
			Show Archived
		</fancycheckbox>

		<div class="namespace" v-for="n in namespaces" :key="`n${n.id}`">
			<h1>
				<span>{{ n.title }}</span>
				<span class="is-archived" v-if="n.isArchived">
					Archived
				</span>
			</h1>

			<div class="lists">
				<template v-for="l in n.lists">
					<router-link
							:to="{ name: 'list.index', params: { listId: l.id} }"
							class="list"
							:key="`l${l.id}`"
							v-if="showArchived ? true : !l.isArchived"
							:style="{
							'background-color': l.hexColor,
							'background-image': typeof backgrounds[l.id] !== 'undefined' ? `url(${backgrounds[l.id]})` : false,
						}"
							:class="{
							'has-light-text': !colorIsDark(l.hexColor),
							'has-background': typeof backgrounds[l.id] !== 'undefined',
						}"
					>
						<div class="is-archived-container">
						<span class="is-archived" v-if="l.isArchived">
							Archived
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
		computed: mapState({
			namespaces(state) {
				return state.namespaces.namespaces.filter(n => this.showArchived ? true : !n.isArchived)
			},
		}),
		created() {
			this.loadBackgroundsForLists()
		},
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
		},
	}
</script>
