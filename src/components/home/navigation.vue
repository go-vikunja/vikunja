<template>
	<aside :class="{'is-active': menuActive}" class="namespace-container">
		<nav class="menu top-menu">
			<router-link :to="{name: 'home'}" class="logo">
				<Logo width="164" height="48"/>
			</router-link>
			<ul class="menu-list">
				<li>
					<router-link :to="{ name: 'home'}" v-shortcut="'g o'">
						<span class="icon">
							<icon icon="calendar"/>
						</span>
						{{ $t('navigation.overview') }}
					</router-link>
				</li>
				<li>
					<router-link :to="{ name: 'tasks.range'}" v-shortcut="'g u'">
						<span class="icon">
							<icon :icon="['far', 'calendar-alt']"/>
						</span>
						{{ $t('navigation.upcoming') }}
					</router-link>
				</li>
				<li>
					<router-link :to="{ name: 'namespaces.index'}" v-shortcut="'g n'">
						<span class="icon">
							<icon icon="layer-group"/>
						</span>
						{{ $t('namespace.title') }}
					</router-link>
				</li>
				<li>
					<router-link :to="{ name: 'labels.index'}" v-shortcut="'g a'">
						<span class="icon">
							<icon icon="tags"/>
						</span>
						{{ $t('label.title') }}
					</router-link>
				</li>
				<li>
					<router-link :to="{ name: 'teams.index'}" v-shortcut="'g m'">
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
						group="namespace-lists"
						@start="() => drag = true"
						@end="saveListPosition"
						handle=".handle"
						:disabled="n.id < 0 || undefined"
						tag="transition-group"
						item-key="id"
						:data-namespace-id="n.id"
						:data-namespace-index="nk"
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

<script setup lang="ts">
import {ref, computed, onMounted, onBeforeMount} from 'vue'
import {useStore} from 'vuex'
import draggable from 'vuedraggable'
import {SortableEvent} from 'sortablejs'

import ListSettingsDropdown from '@/components/list/list-settings-dropdown.vue'
import NamespaceSettingsDropdown from '@/components/namespace/namespace-settings-dropdown.vue'
import PoweredByLink from '@/components/home/PoweredByLink.vue'
import Logo from '@/components/home/Logo.vue'

import {MENU_ACTIVE} from '@/store/mutation-types'
import {calculateItemPosition} from '@/helpers/calculateItemPosition'
import {getNamespaceTitle} from '@/helpers/getNamespaceTitle'
import {useEventListener} from '@vueuse/core'
import NamespaceModel from '@/models/namespace'
import ListModel from '@/models/list'

const drag = ref(false)
const dragOptions = {
	animation: 100,
	ghostClass: 'ghost',
}

const store = useStore()
const currentList = computed(() => store.state.currentList)
const menuActive = computed(() => store.state.menuActive)
const loading = computed(() => store.state.loading && store.state.loadingModule === 'namespaces')


const namespaces = computed(() => {
	return (store.state.namespaces.namespaces as NamespaceModel[]).filter(n => !n.isArchived)
})
const activeLists = computed(() => {
	return namespaces.value.map(({lists}) => {
		return lists?.filter(item => {
			return typeof item !== 'undefined' && !item.isArchived
		})
	})
})

const namespaceTitles = computed(() => {
	return namespaces.value.map((namespace) => getNamespaceTitle(namespace))
})

const namespaceListsCount = computed(() => {
	return namespaces.value.map((_, index) => activeLists.value[index]?.length ?? 0)
})


useEventListener('resize', resize)
onMounted(() => resize())


function toggleFavoriteList(list: ListModel) {
	// The favorites pseudo list is always favorite
	// Archived lists cannot be marked favorite
	if (list.id === -1 || list.isArchived) {
		return
	}
	store.dispatch('lists/toggleListFavorite', list)
}

function resize() {
	// Hide the menu by default on mobile
	store.commit(MENU_ACTIVE, window.innerWidth >= 770)
}

function toggleLists(namespaceId: number) {
	listsVisible.value[namespaceId] = !listsVisible.value[namespaceId]
}

const listsVisible = ref<{ [id: NamespaceModel['id']]: boolean }>({})
// FIXME: async action will be unfinished when component mounts
onBeforeMount(async () => {
	const namespaces = await store.dispatch('namespaces/loadNamespaces') as NamespaceModel[]
	namespaces.forEach(n => {
		if (typeof listsVisible.value[n.id] === 'undefined') {
			listsVisible.value[n.id] = true
		}
	})
})

function updateActiveLists(namespace: NamespaceModel, activeLists: ListModel[]) {
	// This is a bit hacky: since we do have to filter out the archived items from the list
	// for vue draggable updating it is not as simple as replacing it.
	// To work around this, we merge the active lists with the archived ones. Doing so breaks the order
	// because now all archived lists are sorted after the active ones. This is fine because they are sorted 
	// later when showing them anyway, and it makes the merging happening here a lot easier.
	const lists = [
		...activeLists,
		...namespace.lists.filter(l => l.isArchived),
	]

	store.commit('namespaces/setNamespaceById', {
		...namespace,
		lists,
	})
}

const listUpdating = ref<{ [id: NamespaceModel['id']]: boolean }>({})

async function saveListPosition(e: SortableEvent) {
	if (!e.newIndex) return

	const namespaceId = parseInt(e.to.dataset.namespaceId as string)
	const newNamespaceIndex = parseInt(e.to.dataset.namespaceIndex as string)

	const listsActive = activeLists.value[newNamespaceIndex]
	const list = listsActive[e.newIndex]
	const listBefore = listsActive[e.newIndex - 1] ?? null
	const listAfter = listsActive[e.newIndex + 1] ?? null
	listUpdating.value[list.id] = true

	const position = calculateItemPosition(
		listBefore !== null ? listBefore.position : null,
		listAfter !== null ? listAfter.position : null,
	)

	try {
		// create a copy of the list in order to not violate vuex mutations
		await store.dispatch('lists/updateList', {
			...list,
			position,
			namespaceId,
		})
	} finally {
		listUpdating.value[list.id] = false
	}
}
</script>

<style lang="scss" scoped>
$navbar-padding: 2rem;
$vikunja-nav-background: var(--site-background);
$vikunja-nav-color: var(--grey-700);
$vikunja-nav-selected-width: 0.4rem;

.namespace-container {
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
		z-index: 20;
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

			&:hover {
				background: transparent;
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
	width: 2.5rem;
	flex-shrink: 0;
}

.namespaces-list.loader-container.is-loading {
	min-height: calc(100vh - #{$navbar-height + 1.5rem + 1rem + 1.5rem});
}

a.dropdown-item:hover {
	background: var(--dropdown-item-hover-background-color) !important;
}
</style>
