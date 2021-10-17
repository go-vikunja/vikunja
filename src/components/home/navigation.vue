<template>
	<div :class="{'is-active': menuActive}" class="namespace-container">
		<div class="menu top-menu">
			<router-link :to="{name: 'home'}" class="logo">
				<img alt="Vikunja" :src="logoUrl" width="164" height="48"/>
			</router-link>
			<ul class="menu-list">
				<li>
					<router-link :to="{ name: 'home'}">
						<span class="icon">
							<icon icon="calendar"/>
						</span>
						{{ $t('navigation.overview') }}
					</router-link>
				</li>
				<li>
					<router-link :to="{ name: 'tasks.range'}">
						<span class="icon">
							<icon :icon="['far', 'calendar-alt']"/>
						</span>
						{{ $t('navigation.upcoming') }}
					</router-link>
				</li>
				<li>
					<router-link :to="{ name: 'namespaces.index'}">
						<span class="icon">
							<icon icon="layer-group"/>
						</span>
						{{ $t('namespace.title') }}
					</router-link>
				</li>
				<li>
					<router-link :to="{ name: 'labels.index'}">
						<span class="icon">
							<icon icon="tags"/>
						</span>
						{{ $t('label.title') }}
					</router-link>
				</li>
				<li>
					<router-link :to="{ name: 'teams.index'}">
						<span class="icon">
							<icon icon="users"/>
						</span>
						{{ $t('team.title') }}
					</router-link>
				</li>
			</ul>
		</div>

		<aside class="menu namespaces-lists loader-container" :class="{'is-loading': loading}">
			<template v-for="(n, nk) in namespaces" :key="n.id" >
				<div class="namespace-title" :class="{'has-menu': n.id > 0}">
					<span
						@click="toggleLists(n.id)"
						class="menu-label"
						v-tooltip="namespaceTitles[nk]">
						<span class="name">
							<span
								:style="{ backgroundColor: n.hexColor }"
								class="color-bubble"
								v-if="n.hexColor !== ''">
							</span>
							{{ namespaceTitles[nk] }}
						</span>
					</span>
					<a
						class="icon is-small toggle-lists-icon"
						:class="{'active': typeof listsVisible[n.id] !== 'undefined' ? listsVisible[n.id] : true}"
						@click="toggleLists(n.id)"
					>
						<icon icon="chevron-down"/>
					</a>
					<namespace-settings-dropdown :namespace="n" v-if="n.id > 0"/>
				</div>
				<div
					:key="n.id + 'child'"
					class="more-container"
					v-if="typeof listsVisible[n.id] !== 'undefined' ? listsVisible[n.id] : true"
				>
					<!--
						NOTE: a v-model / computed setter is not possible, since the updateActiveLists function
						triggered by the change needs to have access to the current namespace
					--> 
					<draggable
						v-bind="dragOptions"
						:modelValue="activeLists[nk]"
						@update:modelValue="(lists) => updateActiveLists(n, lists)"
						:group="`namespace-${n.id}-lists`"
						@start="() => drag = true"
						@end="e => saveListPosition(e, nk)"
						handle=".handle"
						:disabled="n.id < 0 || null"
						tag="transition-group"
						item-key="id"
						:component-data="{
							type: 'transition',
							tag: 'ul',
							name: !drag ? 'flip-list' : null,
							class: [
								'menu-list can-be-hidden',
								{ 'dragging-disabled': n.id < 0 }
							]
						}"
					>
						<template #item="{element: l}">
							<li
								class="loader-container"
								:class="{'is-loading': listUpdating[l.id]}"
							>
								<router-link
									:to="{ name: 'list.index', params: { listId: l.id} }"
									v-slot="{ href, navigate, isActive }"
									custom
								>
									<a
										@click="navigate"
										:href="href"
										class="list-menu-link"
										:class="{'router-link-exact-active': isActive || currentList?.id === l.id}"
									>
										<span class="icon handle">
											<icon icon="grip-lines"/>
										</span>
										<span
											:style="{ backgroundColor: l.hexColor }"
											class="color-bubble"
											v-if="l.hexColor !== ''">
										</span>
										<span class="list-menu-title">
											{{ getListTitle(l) }}
										</span>
										<span
											:class="{'is-favorite': l.isFavorite}"
											@click.prevent.stop="toggleFavoriteList(l)"
											class="favorite">
											<icon icon="star" v-if="l.isFavorite"/>
											<icon :icon="['far', 'star']" v-else/>
										</span>
									</a>
								</router-link>
								<list-settings-dropdown :list="l" v-if="l.id > 0"/>
								<span class="list-setting-spacer" v-else></span>
							</li>
						</template>
					</draggable>
				</div>
			</template>
		</aside>
		<a class="menu-bottom-link" href="https://vikunja.io" target="_blank" rel="noreferrer noopener nofollow">
			{{ $t('misc.poweredBy') }}
		</a>
	</div>
</template>

<script>
import {mapState} from 'vuex'
import {CURRENT_LIST, MENU_ACTIVE, LOADING, LOADING_MODULE} from '@/store/mutation-types'
import ListSettingsDropdown from '@/components/list/list-settings-dropdown.vue'
import NamespaceSettingsDropdown from '@/components/namespace/namespace-settings-dropdown.vue'
import draggable from 'vuedraggable'
import {calculateItemPosition} from '@/helpers/calculateItemPosition'

import logoUrl from '@/assets/logo-full.svg'

export default {
	name: 'navigation',
	data() {
		return {
			listsVisible: {},
			drag: false,
			dragOptions: {
				animation: 100,
				ghostClass: 'ghost',
			},
			listUpdating: {},
			logoUrl,
		}
	},
	components: {
		ListSettingsDropdown,
		NamespaceSettingsDropdown,
		draggable,
	},
	computed: {
		...mapState({
			namespaces: state => state.namespaces.namespaces.filter(n => !n.isArchived),
			currentList: CURRENT_LIST,
			background: 'background',
			menuActive: MENU_ACTIVE,
			loading: state => state[LOADING] && state[LOADING_MODULE] === 'namespaces',
		}),
		activeLists() {
			return this.namespaces.map(({lists}) => lists?.filter(item => !item.isArchived))
		},
		namespaceTitles() {
			return this.namespaces.map((namespace, index) => {
				const title = this.getNamespaceTitle(namespace)
				return `${title} (${this.activeLists[index]?.length ?? 0})`
			})
		},
	},
	beforeCreate() {
		// FIXME: async action in beforeCreate, might be unfinished when component mounts
		this.$store.dispatch('namespaces/loadNamespaces')
			.then(namespaces => {
				namespaces.forEach(n => {
					if (typeof this.listsVisible[n.id] === 'undefined') {
						this.listsVisible[n.id] = true
					}
				})
			})
	},
	created() {
		window.addEventListener('resize', this.resize)
	},
	mounted() {
		this.resize()
	},
	methods: {
		toggleFavoriteList(list) {
			// The favorites pseudo list is always favorite
			// Archived lists cannot be marked favorite
			if (list.id === -1 || list.isArchived) {
				return
			}
			this.$store.dispatch('lists/toggleListFavorite', list)
		},
		resize() {
			// Hide the menu by default on mobile
			this.$store.commit(MENU_ACTIVE, window.innerWidth >= 770)
		},
		toggleLists(namespaceId) {
			this.listsVisible[namespaceId] = !this.listsVisible[namespaceId]
		},
		updateActiveLists(namespace, activeLists) {
			// this is a bit hacky: since we do have to filter out the archived items from the list
			// for vue draggable updating it is not as simple as replacing it.
			// instead we iterate over the non archived items in the old list and replace them with the ones in their new order
			const lists = namespace.lists.map((item) => {
				if (item.isArchived) {
					return item
				}
				return activeLists.shift()
			})

			const newNamespace = {
				...namespace,
				lists,
			}

			this.$store.commit('namespaces/setNamespaceById', newNamespace)
		},

		async saveListPosition(e, namespaceIndex) {
			const listsActive = this.activeLists[namespaceIndex]
			const list = listsActive[e.newIndex]
			const listBefore = listsActive[e.newIndex - 1] ?? null
			const listAfter = listsActive[e.newIndex + 1] ?? null
			this.listUpdating[list.id] = true

			const position = calculateItemPosition(listBefore !== null ? listBefore.position : null, listAfter !== null ? listAfter.position : null)

			try {
				// create a copy of the list in order to not violate vuex mutations
				await this.$store.dispatch('lists/updateList', {
					...list,
					position,
				})
			} finally {
				this.listUpdating[list.id] = false
			}
		},
	},
}
</script>

<style scoped>
.list-setting-spacer {
	width: 32px;
	flex-shrink: 0;
}
</style>
