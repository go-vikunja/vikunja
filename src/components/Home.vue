<template>
	<div class="content has-text-centered">
		<h2>Hi {{user.infos.username}}!</h2>
		<p>Click on a list or namespace on the left to get started.</p>
		<p v-if="loading">Loading tasks...</p>
		<h3 v-if="tasks && tasks.length > 0">Current tasks</h3>
		<div class="box tasks" v-if="tasks && tasks.length > 0">
			<div @click="gotoList(l.listID)" class="task" v-for="l in tasks" v-bind:key="l.id" v-if="!l.done">
				<label v-bind:for="l.id">
					<div class="fancycheckbox">
						<input @change="markAsDone" type="checkbox" v-bind:id="l.id" v-bind:checked="l.done" style="display: none;" disabled>
						<label  v-bind:for="l.id" class="check">
							<svg width="18px" height="18px" viewBox="0 0 18 18">
								<path d="M1,9 L1,3.5 C1,2 2,1 3.5,1 L14.5,1 C16,1 17,2 17,3.5 L17,14.5 C17,16 16,17 14.5,17 L3.5,17 C2,17 1,16 1,14.5 L1,9 Z"></path>
								<polyline points="1 9 7 14 15 4"></polyline>
							</svg>
						</label>
					</div>
					<span class="tasktext">
						{{l.text}}
						<i v-if="l.dueDate > 0"> - Due on {{formatUnixDate(l.dueDate)}}</i>
					</span>
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
				const cancel = message.setLoading(this)
				HTTP.get(`tasks`, {headers: {'Authorization': 'Bearer ' + localStorage.getItem('token')}})
					.then(response => {
						this.tasks = response.data
						this.tasks.sort(this.sortyByDeadline)
						cancel()
						this.loading = false
					})
					.catch(e => {
						cancel()
						this.loading = false
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