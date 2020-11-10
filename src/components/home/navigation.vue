<template>
	<div :class="{'is-active': menuActive}" class="namespace-container">
		<div class="menu top-menu">
			<router-link :to="{name: 'home'}" class="logo">
				<img alt="Vikunja" src="/images/logo-full.svg"/>
			</router-link>
			<ul class="menu-list">
				<li>
					<router-link :to="{ name: 'home'}">
						<span class="icon">
							<icon icon="calendar"/>
						</span>
						Overview
					</router-link>
				</li>
				<li>
					<router-link :to="{ name: 'tasks.range', params: {type: 'week'}}">
						<span class="icon">
							<icon icon="calendar-week"/>
						</span>
						Next Week
					</router-link>
				</li>
				<li>
					<router-link :to="{ name: 'tasks.range', params: {type: 'month'}}">
						<span class="icon">
							<icon :icon="['far', 'calendar-alt']"/>
						</span>
						Next Month
					</router-link>
				</li>
				<li>
					<router-link :to="{ name: 'teams.index'}">
						<span class="icon">
							<icon icon="users"/>
						</span>
						Teams
					</router-link>
				</li>
				<li>
					<router-link :to="{ name: 'namespaces.index'}">
						<span class="icon">
							<icon icon="layer-group"/>
						</span>
						Namespaces & Lists
					</router-link>
				</li>
				<li>
					<router-link :to="{ name: 'labels.index'}">
						<span class="icon">
							<icon icon="tags"/>
						</span>
						Labels
					</router-link>
				</li>
			</ul>
		</div>

		<aside class="menu namespaces-lists">
			<template v-for="n in namespaces">
				<div :key="n.id">
					<router-link
						:to="{name: 'namespace.edit', params: {id: n.id} }"
						class="nsettings"
						v-if="n.id > 0"
						v-tooltip="'Settings'">
						<span class="icon">
							<icon icon="cog"/>
						</span>
					</router-link>
					<router-link
						:key="n.id + 'list.create'"
						:to="{ name: 'list.create', params: { id: n.id} }"
						class="nsettings"
						v-if="n.id > 0"
						v-tooltip="'Add a new list in the ' + n.title + ' namespace'">
						<span class="icon">
							<icon icon="plus"/>
						</span>
					</router-link>
					<label
						:for="n.id + 'checker'"
						class="menu-label"
						v-tooltip="n.title + ' (' + n.lists.length + ')'">
						<span class="name">
							<span
								:style="{ backgroundColor: n.hexColor }"
								class="color-bubble"
								v-if="n.hexColor !== ''">
							</span>
							{{ n.title }} ({{ n.lists.length }})
						</span>
					</label>
				</div>
				<input
					:id="n.id + 'checker'"
					:key="n.id + 'checker'"
					checked="checked"
					class="checkinput"
					type="checkbox"/>
				<div :key="n.id + 'child'" class="more-container">
					<ul class="menu-list can-be-hidden">
						<template v-for="l in n.lists">
							<!-- This is a bit ugly but vue wouldn't want to let me filter this - probably because the lists
									are nested inside of the namespaces makes it a lot harder.-->
							<li :key="l.id" v-if="!l.isArchived">
								<router-link
									class="list-menu-link"
									:class="{'router-link-exact-active': currentList.id === l.id}"
									:to="{ name: 'list.index', params: { listId: l.id} }"
									tag="span"
								>
									<span
										:style="{ backgroundColor: l.hexColor }"
										class="color-bubble"
										v-if="l.hexColor !== ''">
									</span>
									<span class="list-menu-title">
										{{ l.title }}
									</span>
									<span
										:class="{'is-favorite': l.isFavorite}"
										@click.stop="toggleFavoriteList(l)"
										class="favorite">
										<icon icon="star" v-if="l.isFavorite"/>
										<icon :icon="['far', 'star']" v-else/>
									</span>
								</router-link>
							</li>
						</template>
					</ul>
					<label :for="n.id + 'checker'" class="hidden-hint">
						Show hidden lists ({{ n.lists.length }})...
					</label>
				</div>
			</template>
		</aside>
		<a class="menu-bottom-link" href="https://vikunja.io" target="_blank">Powered by Vikunja</a>
	</div>
</template>

<script>
import {mapState} from 'vuex'
import {CURRENT_LIST, IS_FULLPAGE, MENU_ACTIVE} from '@/store/mutation-types'

export default {
	name: 'navigation',
	computed: mapState({
		fullpage: IS_FULLPAGE,
		namespaces(state) {
			return state.namespaces.namespaces.filter(n => !n.isArchived)
		},
		currentList: CURRENT_LIST,
		background: 'background',
		menuActive: MENU_ACTIVE,
	}),
	beforeCreate() {
		this.$store.dispatch('namespaces/loadNamespaces')
	},
	created() {
		// Hide the menu by default on mobile
		if (window.innerWidth < 770) {
			this.$store.commit(MENU_ACTIVE, false)
		}
	},
	methods: {
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
