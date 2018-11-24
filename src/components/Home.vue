<template>
	<div class="content has-text-centered">
		<h2>Hi {{user.infos.username}}!</h2>
		<p>Click on a list or namespace on the left to get started.</p>
		<p v-if="loading">Loading tasks...</p>
		<h3 v-if="tasks && tasks.length > 0">Current tasks</h3>
		<div class="box tasks" v-if="tasks && tasks.length > 0">
			<div @click="gotoList(l.listID)" class="task" v-for="l in tasks" v-bind:key="l.id" v-if="!l.done">
				<label v-bind:for="l.id">
					<input type="checkbox" v-bind:id="l.id" disabled>
					{{l.text}}
					<i v-if="l.dueDate > 0"> - Due on {{formatUnixDate(l.dueDate)}}</i>
				</label>
			</div>
		</div>
	</div>
</template>

<script>
    import auth from '../auth'
    import router from '../router'
	import {HTTP} from '../http-common'
	import message from '../message'

    export default {
        name: "Home",
        data() {
            return {
                user: auth.user,
				loading: false,
				tasks: []
            }
        },
        beforeMount() {
            // Check if the user is already logged in, if so, redirect him to the homepage
			if (!auth.user.authenticated) {
                router.push({name: 'login'})
            }
        },
		created() {
			if (auth.user.authenticated) {
				this.loadPendingTasks()
			}
		},
        methods: {
            logout() {
                auth.logout()
            },
			loadPendingTasks() {
				this.loading = true
				HTTP.get(`tasks`, {headers: {'Authorization': 'Bearer ' + localStorage.getItem('token')}})
					.then(response => {
						this.tasks = response.data
						this.tasks.sort(this.sortyByDeadline)
						this.loading = false
					})
					.catch(e => {
						this.handleError(e)
					})
            },
			formatUnixDate(dateUnix) {
				return (new Date(dateUnix * 1000)).toLocaleString()
			},
			sortyByDeadline(a, b) {
				return ((a.dueDate > b.dueDate) ? -1 : ((a.dueDate < b.dueDate) ? 1 : 0));
			},
			gotoList(lid) {
				router.push({name: 'showList', params: {id: lid}})
			},
			handleError(e) {
				this.loading = false
				message.error(e, this)
			}
        },
    }
</script>

<style scoped>
h3{
	text-align: left;
}
</style>