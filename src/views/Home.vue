<template>
	<div class="content has-text-centered">
		<h2>
			{{ $t(`home.welcome${welcome}`, {username: userInfo.name !== '' ? userInfo.name : userInfo.username}) }}!
		</h2>
		<add-task
			:listId="defaultListId"
			@taskAdded="updateTaskList"
			class="is-max-width-desktop"
		/>
		<template v-if="!hasTasks && !loading">
			<template v-if="defaultNamespaceId > 0">
				<p class="mt-4">{{ $t('home.list.newText') }}</p>
				<x-button
					:to="{ name: 'list.create', params: { id: defaultNamespaceId } }"
					:shadow="false"
					class="ml-2"
				>
					{{ $t('home.list.new') }}
				</x-button>
			</template>
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
		<div v-if="listHistory.length > 0" class="is-max-width-desktop has-text-left">
			<h3>{{ $t('home.lastViewed') }}</h3>
			<div class="is-flex list-cards-wrapper-2-rows">
				<list-card
					v-for="(l, k) in listHistory"
					:key="`l${k}`"
					:list="l"
					:background-resolver="() => null"
				/>
			</div>
		</div>
		<ShowTasks :show-all="true" v-if="hasLists" :key="showTasksKey"/>
	</div>
</template>

<script>
import {mapState} from 'vuex'
import ShowTasks from './tasks/ShowTasks.vue'
import {getHistory} from '../modules/listHistory'
import ListCard from '@/components/list/partials/list-card.vue'
import AddTask from '../components/tasks/add-task.vue'
import {LOADING, LOADING_MODULE} from '../store/mutation-types'

export default {
	name: 'Home',
	components: {
		ListCard,
		ShowTasks,
		AddTask,
	},
	data() {
		return {
			currentDate: new Date(),
			tasks: [],
			showTasksKey: 0,
		}
	},
	computed: {
		welcome() {
			const now = new Date()

			if (now.getHours() < 5) {
				return 'Night'
			}

			if (now.getHours() < 11) {
				return 'Morning'
			}

			if (now.getHours() < 18) {
				return 'Day'
			}

			if (now.getHours() < 23) {
				return 'Evening'
			}

			return 'Night'
		},
		listHistory() {
			const history = getHistory()
			return history.map(l => {
				return this.$store.getters['lists/getListById'](l.id)
			}).filter(l => l !== null)
		},
		...mapState({
			migratorsEnabled: state =>
				state.config.availableMigrators !== null &&
				state.config.availableMigrators.length > 0,
			authenticated: state => state.auth.authenticated,
			userInfo: state => state.auth.info,
			hasTasks: state => state.hasTasks,
			defaultListId: state => state.auth.defaultListId,
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
			loading: state => state[LOADING] && state[LOADING_MODULE] === 'tasks',
		}),
	},
	methods: {
		// This is to reload the tasks list after adding a new task through the global task add.
		// FIXME: Should use vuex (somehow?)
		updateTaskList() {
			this.showTasksKey++
		},
	},
}
</script>
