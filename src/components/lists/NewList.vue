<template>
	<div class="fullpage">
		<a class="close" @click="back()">
			<icon :icon="['far', 'times-circle']">
			</icon>
		</a>
		<h3>Create a new list</h3>
		<form @submit.prevent="newList" @keyup.esc="back()">
			<div class="field is-grouped">
				<p class="control is-expanded" :class="{ 'is-loading': loading}">
					<input v-focus class="input" :class="{ 'disabled': loading}" v-model="list.title" type="text" placeholder="The list's name goes here...">
				</p>
				<p class="control">
					<button type="submit" class="button is-success noshadow">
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
		created() {
			this.$parent.setFullPage();
		},
        methods: {
            newList() {
				const cancel = message.setLoading(this)

                HTTP.put(`namespaces/` + this.$route.params.id + `/lists`, this.list, {headers: {'Authorization': 'Bearer ' + localStorage.getItem('token')}})
                    .then(response => {
						this.$parent.loadNamespaces()
						this.handleSuccess({message: 'The list was successfully created.'})
						cancel()
						router.push({name: 'showList', params: {id: response.data.id}})
                    })
                    .catch(e => {
                        cancel()
						this.handleError(e)
                    })
            },
			back() {
				router.go(-1)
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