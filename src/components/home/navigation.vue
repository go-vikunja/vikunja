<template>
	<div :class="{'is-active': menuActive}" class="namespace-container">
		<div class="menu top-menu">
			<router-link :to="{name: 'home'}" class="logo">
				<img alt="Vikunja" src="/images/logo-full.svg" width="164" height="48"/>
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
			<template v-for="(n, nk) in namespaces">
				<div :key="n.id" class="namespace-title" :class="{'has-menu': n.id > 0}">
					<span
						@click="toggleLists(n.id)"
						class="menu-label"
						v-tooltip="getNamespaceTitle(n) + ' (' + n.lists.filter(l => !l.isArchived).length + ')'">
						<span class="name">
							<span
								:style="{ backgroundColor: n.hexColor }"
								class="color-bubble"
								v-if="n.hexColor !== ''">
							</span>
							{{ getNamespaceTitle(n) }} ({{ n.lists.filter(l => !l.isArchived).length }})
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
					<draggable
						v-model="n.lists"
						:group="`namespace-${n.id}-lists`"
						@start="() => drag = true"
						@end="e => saveListPosition(e, nk)"
						v-bind="dragOptions"
						handle=".handle"
						:disabled="n.id < 0"
						:class="{'dragging-disabled': n.id < 0}"
					>
						<transition-group
							type="transition"
							:name="!drag ? 'flip-list' : null"
							tag="ul"
							class="menu-list can-be-hidden"
						>
							<!-- eslint-disable vue/no-use-v-if-with-v-for,vue/no-confusing-v-for-v-if -->
							<li
								v-for="l in n.lists"
								:key="l.id"
								v-if="!l.isArchived"
								class="loader-container"
								:class="{'is-loading': listUpdating[l.id]}"
							>
								<router-link
									class="list-menu-link"
									:class="{'router-link-exact-active': currentList.id === l.id}"
									:to="{ name: 'list.index', params: { listId: l.id} }"
									tag="span"
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
										@click.stop="toggleFavoriteList(l)"
										class="favorite">
										<icon icon="star" v-if="l.isFavorite"/>
										<icon :icon="['far', 'star']" v-else/>
									</span>
								</router-link>
								<list-settings-dropdown :list="l" v-if="l.id > 0"/>
								<span class="list-setting-spacer" v-else></span>
							</li>
						</transition-group>
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
		}
	},
	components: {
		ListSettingsDropdown,
		NamespaceSettingsDropdown,
		draggable,
	},
	computed: mapState({
		namespaces: state => state.namespaces.namespaces.filter(n => !n.isArchived),
		currentList: CURRENT_LIST,
		background: 'background',
		menuActive: MENU_ACTIVE,
		loading: state => state[LOADING] && state[LOADING_MODULE] === 'namespaces',
	}),
	beforeCreate() {
		this.$store.dispatch('namespaces/loadNamespaces')
			.then(namespaces => {
				namespaces.forEach(n => {
					if (typeof this.listsVisible[n.id] === 'undefined') {
						this.$set(this.listsVisible, n.id, true)
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
				.catch(e => this.error(e))
		},
		resize() {
			// Hide the menu by default on mobile
			if (window.innerWidth < 770) {
				this.$store.commit(MENU_ACTIVE, false)
			} else {
				this.$store.commit(MENU_ACTIVE, true)
			}
		},
		toggleLists(namespaceId) {
			this.$set(this.listsVisible, namespaceId, !this.listsVisible[namespaceId] ?? false)
		},
		saveListPosition(e, namespaceIndex) {
			const listsFiltered = this.namespaces[namespaceIndex].lists.filter(l => !l.isArchived)
			const list = listsFiltered[e.newIndex]
			const listBefore = listsFiltered[e.newIndex - 1] ?? null
			const listAfter = listsFiltered[e.newIndex + 1] ?? null
			this.$set(this.listUpdating, list.id, true)

			list.position = calculateItemPosition(listBefore !== null ? listBefore.position : null, listAfter !== null ? listAfter.position : null)

			this.$store.dispatch('lists/updateList', list)
				.catch(e => {
					this.error(e)
				})
				.finally(() => {
					this.$set(this.listUpdating, list.id, false)
				})
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
