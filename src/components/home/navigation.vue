<template>
	<aside :class="{'is-active': menuActive}" class="namespace-container">
		<nav class="menu top-menu">
			<router-link :to="{name: 'home'}" class="logo">
				<Logo width="164" height="48"/>
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
		</nav>

		<nav class="menu namespaces-lists loader-container is-loading-small" :class="{'is-loading': loading}">
			<template v-for="(n, nk) in namespaces" :key="n.id">
				<div class="namespace-title" :class="{'has-menu': n.id > 0}">
					<span
						@click="toggleLists(n.id)"
						class="menu-label"
						v-tooltip="namespaceTitles[nk]"
					>
						<span
							v-if="n.hexColor !== ''"
							:style="{ backgroundColor: n.hexColor }"
							class="color-bubble"
						/>
						<span class="name">
							{{ namespaceTitles[nk] }}
						</span>
						<a
							class="icon is-small toggle-lists-icon pl-2"
							:class="{'active': typeof listsVisible[n.id] !== 'undefined' ? listsVisible[n.id] : true}"
							@click="toggleLists(n.id)"
						>
							<icon icon="chevron-down"/>
						</a>
						<span class="count" :class="{'ml-2 mr-0': n.id > 0}">
							({{ namespaceListsCount[nk] }})
						</span>
					</span>
					<namespace-settings-dropdown :namespace="n" v-if="n.id > 0"/>
				</div>
				<div
					v-if="listsVisible[n.id] ?? true"
					:key="n.id + 'child'"
					class="more-container"
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
								class="loader-container is-loading-small"
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
											<icon :icon="l.isFavorite ? 'star' : ['far', 'star']"/>
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
		</nav>
		<PoweredByLink/>
	</aside>
</template>

<script>
import {mapState} from 'vuex'
import draggable from 'vuedraggable'

import ListSettingsDropdown from '@/components/list/list-settings-dropdown.vue'
import NamespaceSettingsDropdown from '@/components/namespace/namespace-settings-dropdown.vue'
import PoweredByLink from '@/components/home/PoweredByLink.vue'
import Logo from '@/components/home/Logo.vue'

import {CURRENT_LIST, MENU_ACTIVE, LOADING, LOADING_MODULE} from '@/store/mutation-types'
import {calculateItemPosition} from '@/helpers/calculateItemPosition'


export default {
	name: 'navigation',

	components: {
		ListSettingsDropdown,
		NamespaceSettingsDropdown,
		draggable,
		Logo,
		PoweredByLink,
	},

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
			return this.namespaces.map((namespace) => this.getNamespaceTitle(namespace))
		},
		namespaceListsCount() {
			return this.namespaces.map((_, index) => this.activeLists[index]?.length ?? 0)
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

<style lang="scss" scoped>
$navbar-padding: 2rem;
$vikunja-nav-background: var(--site-background);
$vikunja-nav-color: var(--grey-700);
$vikunja-nav-selected-width: 0.4rem;

.namespace-container {
	z-index: 6;
	background: $vikunja-nav-background;
	color: $vikunja-nav-color;
	padding: 0 0 1rem;
	transition: transform $transition-duration ease-in;
	position: fixed;
	top: $navbar-height;
	bottom: 0;
	left: 0;
	transform: translateX(-100%);
	overflow-x: auto;
	width: $navbar-width;


	@media screen and (max-width: $tablet) {
		top: 0;
		width: 70vw;
	}

	&.is-active {
		transform: translateX(0);
		transition: transform $transition-duration ease-out;
	}

	.menu {
		.menu-label {
			font-size: 1rem;
			font-weight: 700;
			font-weight: bold;
			font-family: $vikunja-font;
			color: $vikunja-nav-color;
			font-weight: 500;
			min-height: 2.5rem;
			padding-top: 0;
			padding-left: $navbar-padding;

			overflow: hidden;
		}

		.menu-label,
		.menu-list span.list-menu-link,
		.menu-list a {
			display: flex;
			align-items: center;
			justify-content: space-between;
			cursor: pointer;

			.list-menu-title {
				overflow: hidden;
				text-overflow: ellipsis;
				width: 100%;
			}

			.color-bubble {
				height: 12px;
				flex: 0 0 12px;
			}

			.favorite {
				margin-left: .25rem;
				transition: opacity $transition, color $transition;
				opacity: 0;

				&:hover {
					color: var(--warning);
				}

				&.is-favorite {
					opacity: 1;
					color: var(--warning);
				}
			}

			&:hover .favorite {
				opacity: 1;
			}
		}

		.menu-label {
			.color-bubble {
				width: 14px;
				height: 14px;
				flex-basis: auto;
			}

			.is-archived {
				min-width: 85px;
			}
		}

		.namespace-title {
			display: flex;
			align-items: center;
			justify-content: space-between;

			.menu-label {
				margin-bottom: 0;
				flex: 1 1 auto;

				.name {
					overflow: hidden;
					text-overflow: ellipsis;
					white-space: nowrap;
					margin-right: auto;
				}

				.count {
					color: var(--grey-500);
					margin-right: .5rem;
				}
			}

			a:not(.dropdown-item) {
				color: $vikunja-nav-color;
				padding: 0 .25rem;
			}

			:deep(.dropdown-trigger) {
				padding: .5rem;
				cursor: pointer;
			}

			.toggle-lists-icon {
				svg {
					transition: all $transition;
					transform: rotate(90deg);
					opacity: 1;
				}

				&.active svg {
					transform: rotate(0deg);
					opacity: 0;
				}
			}

			&:hover .toggle-lists-icon svg {
				opacity: 1;
			}

			&:not(.has-menu) .toggle-lists-icon {
				padding-right: 1rem;
			}
		}

		.menu-label,
		.nsettings,
		.menu-list span.list-menu-link,
		.menu-list a {
			color: $vikunja-nav-color;
		}

		.menu-list {
			li {
				height: 44px;
				display: flex;
				align-items: center;

				&:hover {
					background: var(--white);
				}

				:deep(.dropdown-trigger) {
					opacity: 0;
					padding: .5rem;
					cursor: pointer;
					transition: $transition;
				}

				&:hover :deep(.dropdown-trigger) {
					opacity: 1;
				}
			}

			.flip-list-move {
				transition: transform $transition-duration;
			}

			.ghost {
				background: var(--grey-200);

				* {
					opacity: 0;
				}
			}

			a:hover {
				background: transparent;
			}

			span.list-menu-link, li > a {
				padding: 0.75rem .5rem 0.75rem ($navbar-padding * 1.5 - 1.75rem);
				transition: all 0.2s ease;

				border-radius: 0;
				white-space: nowrap;
				text-overflow: ellipsis;
				overflow: hidden;
				width: 100%;
				border-left: $vikunja-nav-selected-width solid transparent;

				.icon {
					height: 1rem;
					vertical-align: middle;
					padding-right: 0.5rem;

					&.handle {
						opacity: 0;
						transition: opacity $transition;
						margin-right: .25rem;
						cursor: grab;
					}
				}

				&:hover .icon.handle {
					opacity: 1;
				}

				&.router-link-exact-active {
					color: var(--primary);
					border-left: $vikunja-nav-selected-width solid var(--primary);

					.icon {
						color: var(--primary);
					}
				}

				&:hover {
					border-left: $vikunja-nav-selected-width solid var(--primary);
				}
			}
		}

		.logo {
			display: block;

			padding-left: 2rem;
			margin-right: 1rem;

			@media screen and (min-width: $tablet) {
				display: none;
			}
		}

		&.namespaces-lists {
			padding-top: math.div($navbar-padding, 2);
		}

		.icon {
			color: var(--grey-400) !important;
		}
	}

	.top-menu {
		margin-top: math.div($navbar-padding, 2);

		.menu-list {
			li {
				font-weight: 500;
				font-family: $vikunja-font;
			}

			span.list-menu-link, li > a {
				padding-left: 2rem;
				display: inline-block;

				.icon {
					padding-bottom: .25rem;
				}
			}
		}
	}
}

.list-setting-spacer {
	width: 32px;
	flex-shrink: 0;
}

.namespaces-list.loader-container.is-loading {
	min-height: calc(100vh - #{$navbar-height + 1.5rem + 1rem + 1.5rem});
}
</style>
