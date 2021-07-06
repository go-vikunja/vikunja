<template>
	<div class="content has-text-centered">
		<h2>
			{{ $t(`home.welcome${welcome}`, {username: userInfo.name !== '' ? userInfo.name : userInfo.username}) }}!
		</h2>
		<template v-if="!hasTasks">
			<p>{{ $t('home.list.newText') }}</p>
			<x-button
				:to="{name: 'list.create', params: { id: defaultNamespaceId }}"
				:shadow="false"
				class="ml-2"
				v-if="defaultNamespaceId > 0"
			>
				{{ $t('home.list.new') }}
			</x-button>
			<p class="mt-4" v-if="migratorsEnabled">
				{{ $t('home.list.importText') }}
			</p>
			<x-button
				v-if="migratorsEnabled"
				:to="{ name: 'migrate.start' }"
				:shadow="false">
				{{ $t('home.list.import') }}
			</x-button>
		</template>
		<ShowTasks :show-all="true" v-if="hasLists"/>
	</div>
</template>

<script>
import {mapState} from 'vuex'
import ShowTasks from './tasks/ShowTasks'

export default {
	name: 'Home',
	components: {
		ShowTasks,
	},
	data() {
		return {
			loading: false,
			currentDate: new Date(),
			tasks: [],
		}
	},
	computed: {
		welcome() {
			const now = new Date()

			if (now.getHours() < 5) {
				return 'Night'
			}

			if(now.getHours() < 11) {
				return 'Morning'
			}

			if(now.getHours() < 18) {
				return 'Day'
			}

			if(now.getHours() < 23) {
				return 'Evening'
			}

			return 'Night'
		},
		...mapState({
			migratorsEnabled: state => state.config.availableMigrators !== null && state.config.availableMigrators.length > 0,
			authenticated: state => state.auth.authenticated,
			userInfo: state => state.auth.info,
			hasTasks: state => state.hasTasks,
			defaultNamespaceId: state => {
				if (state.namespaces.namespaces.length === 0) {
					return 0
				}

				return state.namespaces.namespaces[0].id
			},
			hasLists: state => {
				if (state.namespaces.namespaces.length === 0) {
					return false
				}

				return state.namespaces.namespaces[0].lists.length > 0
			},
		}),
	}
}
</script>
