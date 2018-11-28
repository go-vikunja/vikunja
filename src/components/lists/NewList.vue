<template>
	<div class="content">
		<h3>Create a new list</h3>
		<form @submit.prevent="newList">
			<div class="field is-grouped">
				<p class="control has-icons-left is-expanded" v-bind:class="{ 'is-loading': loading}">
					<input class="input" v-bind:class="{ 'disabled': loading}" v-model="list.title" type="text" placeholder="The list's name goes here...">
					<span class="icon is-small is-left">
						<icon icon="list-ol"/>
					</span>
				</p>
				<p class="control">
					<button type="submit" class="button is-success">
						<span class="icon is-small">
							<icon icon="plus"/>
						</span>
						Add
					</button>
				</p>
			</div>
		</form>
	</div>
</template>

<script>
    import auth from '../../auth'
    import router from '../../router'
    import {HTTP} from '../../http-common'
    import message from '../../message'

    export default {
        name: "NewList",
        data() {
            return {
                list: {title: ''},
                error: '',
                loading: false
            }
        },
        beforeMount() {
            // Check if the user is already logged in, if so, redirect him to the homepage
            if (!auth.user.authenticated) {
                router.push({name: 'home'})
            }
        },
        methods: {
            newList() {
				const cancel = message.setLoading(this)

                HTTP.put(`namespaces/` + this.$route.params.id + `/lists`, this.list, {headers: {'Authorization': 'Bearer ' + localStorage.getItem('token')}})
                    .then(() => {
						this.$parent.loadNamespaces()
						this.handleSuccess({message: 'The list was successfully created.'})
						cancel()
                    })
                    .catch(e => {
                        cancel()
						this.handleError(e)
                    })
            },
            handleError(e) {
                message.error(e, this)
            },
            handleSuccess(e) {
                message.success(e, this)
            }
        }
    }
</script>

<style scoped>

</style>