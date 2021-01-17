<template>
	<div class="content has-text-centered">
		<h2>
			Hi {{ userInfo.name !== '' ? userInfo.name : userInfo.username }}!
		</h2>
		<template v-if="!hasTasks">
			<p>You can create a new list for your new tasks:</p>
			<x-button
				:to="{name: 'list.create', params: { id: defaultNamespaceId }}"
				:shadow="false"
				class="ml-2"
				v-if="defaultNamespaceId > 0"
			>
				Create a new list
			</x-button>
			<p class="mt-4" v-if="migratorsEnabled">
				Or import your lists and tasks from other services into Vikunja:
			</p>
			<x-button :to="{ name: 'migrate.start' }" :shadow="false">
				Import your data into Vikunja
			</x-button>
		</template>
		<ShowTasks :show-all="true" v-if="hasLists"/>
	</div>
</template>

<script>
import { mapState } from 'vuex'
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
	computed: mapState({
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
</script>
