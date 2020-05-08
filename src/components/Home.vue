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
		<TaskOverview :show-all="true"/>
	</div>
</template>

<script>
	import router from '../router'
	import {mapState} from 'vuex'

	export default {
		name: "Home",
		data() {
			return {
				loading: false,
				currentDate: new Date(),
				tasks: []
			}
		},
		beforeMount() {
			// Check if the user is already logged in, if so, redirect him to the homepage
			if (!this.authenticated) {
				router.push({name: 'login'})
			}
		},
		computed: mapState({
			migratorsEnabled: state => state.config.availableMigrators !== null && state.config.availableMigrators.length > 0,
			authenticated: state => state.auth.authenticated,
			userInfo: state => state.auth.info,
		}),
	}
</script>
