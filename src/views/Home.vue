<template>
	<div class="content has-text-centered">
		<h2>Hi {{ userInfo.name !== '' ? userInfo.name : userInfo.username }}!</h2>
		<template v-if="!hasTasks">
			<p>Click on a list or namespace on the left to get started.</p>
			<router-link
				:to="{name: 'migrate.start'}"
				class="button is-primary has-no-shadow"
				v-if="migratorsEnabled"
			>
				Import your data into Vikunja
			</router-link>
		</template>
		<ShowTasks :show-all="true"/>
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
	computed: mapState({
		migratorsEnabled: state => state.config.availableMigrators !== null && state.config.availableMigrators.length > 0,
		authenticated: state => state.auth.authenticated,
		userInfo: state => state.auth.info,
		hasTasks: state => state.hasTasks,
	}),
}
</script>
