<template>
	<div class="content has-text-centered">
		<h2>Hi {{userInfo.username}}!</h2>
		<p>Click on a list or namespace on the left to get started.</p>
		<router-link
				class="button is-primary is-right noshadow is-outlined"
				:to="{name: 'migrateStart'}"
				v-if="migratorsEnabled"
		>
			Import your data into Vikunja
		</router-link>
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
				tasks: []
			}
		},
		computed: mapState({
			migratorsEnabled: state => state.config.availableMigrators !== null && state.config.availableMigrators.length > 0,
			authenticated: state => state.auth.authenticated,
			userInfo: state => state.auth.info,
		}),
	}
</script>
