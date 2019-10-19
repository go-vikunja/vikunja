<template>
    <form @submit.prevent="editTaskSubmit()">
        <div class="field">
            <label class="label" for="tasktext">Task Text</label>
            <div class="control">
                <input v-focus :class="{ 'disabled': taskService.loading}" :disabled="taskService.loading" class="input" type="text" id="tasktext" placeholder="The task text is here..." v-model="taskEditTask.text">
            </div>
        </div>
        <div class="field">
            <label class="label" for="taskdescription">Description</label>
            <div class="control">
                <textarea :class="{ 'disabled': taskService.loading}" :disabled="taskService.loading" class="textarea" placeholder="The tasks description goes here..." id="taskdescription" v-model="taskEditTask.description"></textarea>
            </div>
        </div>

        <b>Reminder Dates</b>
        <div class="reminder-input" :class="{ 'overdue': (r < nowUnix && index !== (taskEditTask.reminderDates.length - 1))}" v-for="(r, index) in taskEditTask.reminderDates" :key="index">
            <flat-pickr
                    :class="{ 'disabled': taskService.loading}"
                    :disabled="taskService.loading"
                    :v-model="taskEditTask.reminderDates"
                    :config="flatPickerConfig"
                    :id="'taskreminderdate' + index"
                    :value="r"
                    :data-index="index"
                    placeholder="Add a new reminder...">
            </flat-pickr>
            <a v-if="index !== (taskEditTask.reminderDates.length - 1)" @click="removeReminderByIndex(index)"><icon icon="times"></icon></a>
        </div>

        <div class="field">
            <label class="label" for="taskduedate">Due Date</label>
            <div class="control">
                <flat-pickr
                        :class="{ 'disabled': taskService.loading}"
                        class="input"
                        :disabled="taskService.loading"
                        v-model="taskEditTask.dueDate"
                        :config="flatPickerConfig"
                        id="taskduedate"
                        placeholder="The tasks due date is here...">
                </flat-pickr>
            </div>
        </div>

        <div class="field">
            <label class="label" for="">Duration</label>
            <div class="control columns">
                <div class="column">
                    <flat-pickr
                            :class="{ 'disabled': taskService.loading}"
                            class="input"
                            :disabled="taskService.loading"
                            v-model="taskEditTask.startDate"
                            :config="flatPickerConfig"
                            id="taskduedate"
                            placeholder="Start date">
                    </flat-pickr>
                </div>
                <div class="column">
                    <flat-pickr
                            :class="{ 'disabled': taskService.loading}"
                            class="input"
                            :disabled="taskService.loading"
                            v-model="taskEditTask.endDate"
                            :config="flatPickerConfig"
                            id="taskduedate"
                            placeholder="End date">
                    </flat-pickr>
                </div>
            </div>
        </div>

        <div class="field">
            <label class="label" for="">Repeat after</label>
            <div class="control repeat-after-input columns">
                <div class="column">
                    <input class="input" placeholder="Specify an amount..." v-model="taskEditTask.repeatAfter.amount"/>
                </div>
                <div class="column is-3">
                    <div class="select">
                        <select v-model="taskEditTask.repeatAfter.type">
                            <option value="hours">Hours</option>
                            <option value="days">Days</option>
                            <option value="weeks">Weeks</option>
                            <option value="months">Months</option>
                            <option value="years">Years</option>
                        </select>
                    </div>
                </div>
            </div>
        </div>

        <div class="field">
            <label class="label" for="">Priority</label>
            <div class="control priority-select">
                <div class="select">
                    <select v-model="taskEditTask.priority">
                        <option :value="priorities.UNSET">Unset</option>
                        <option :value="priorities.LOW">Low</option>
                        <option :value="priorities.MEDIUM">Medium</option>
                        <option :value="priorities.HIGH">High</option>
                        <option :value="priorities.URGENT">Urgent</option>
                        <option :value="priorities.DO_NOW">DO NOW</option>
                    </select>
                </div>
            </div>
        </div>

        <div class="field">
            <label class="label">Percent Done</label>
            <div class="control">
                <div class="select">
                    <select v-model.number="taskEditTask.percentDone">
                        <option value="0">0%</option>
                        <option value="0.1">10%</option>
                        <option value="0.2">20%</option>
                        <option value="0.3">30%</option>
                        <option value="0.4">40%</option>
                        <option value="0.5">50%</option>
                        <option value="0.6">60%</option>
                        <option value="0.7">70%</option>
                        <option value="0.8">80%</option>
                        <option value="0.9">90%</option>
                        <option value="1">100%</option>
                    </select>
                </div>
            </div>
        </div>

        <div class="field">
            <label class="label">Color</label>
            <div class="control">
                <verte
                        v-model="taskEditTask.hexColor"
                        menuPosition="top"
                        picker="square"
                        model="hex"
                        :enableAlpha="false"
                        :rgbSliders="true">
                </verte>
            </div>
        </div>

        <div class="field">
            <label class="label" for="">Assignees</label>
            <ul class="assingees">
                <li v-for="(a, index) in taskEditTask.assignees" :key="a.id">
                    {{a.username}}
                    <a @click="deleteAssigneeByIndex(index)"><icon icon="times"/></a>
                </li>
            </ul>
        </div>

        <div class="field has-addons">
            <div class="control is-expanded">
                <multiselect
                        v-model="newAssignee"
                        :options="foundUsers"
                        :multiple="false"
                        :searchable="true"
                        :loading="listUserService.loading"
                        :internal-search="true"
                        @search-change="findUser"
                        placeholder="Type to search"
                        label="username"
                        track-by="id">
                    <template slot="clear" slot-scope="props">
                        <div class="multiselect__clear" v-if="newAssignee !== null && newAssignee.id !== 0" @mousedown.prevent.stop="clearAllFoundUsers(props.search)"></div>
                    </template>
                    <span slot="noResult">Oops! No user found. Consider changing the search query.</span>
                </multiselect>
            </div>
            <div class="control">
                <a @click="addAssignee" class="button is-primary fullheight">
                                    <span class="icon is-small">
                                        <icon icon="plus"/>
                                    </span>
                </a>
            </div>
        </div>

        <div class="field">
            <label class="label">Labels</label>
            <div class="control">
                <multiselect
                        :multiple="true"
                        :close-on-select="false"
                        :clear-on-select="true"
                        :options-limit="300"
                        :hide-selected="true"
                        v-model="taskEditTask.labels"
                        :options="foundLabels"
                        :searchable="true"
                        :loading="labelService.loading || labelTaskService.loading"
                        :internal-search="true"
                        @search-change="findLabel"
                        @select="addLabel"
                        placeholder="Type to search"
                        label="title"
                        track-by="id"
                        :taggable="true"
                        @tag="createAndAddLabel"
                        tag-placeholder="Add this as new label"
                >
                    <template slot="tag" slot-scope="{ option, remove }">
                                        <span class="tag" :style="{'background': option.hex_color, 'color': option.textColor}">
                                            <span>{{ option.title }}</span>
                                            <a class="delete is-small" @click="removeLabel(option)"></a>
                                        </span>
                    </template>
                    <template slot="clear" slot-scope="props">
                        <div class="multiselect__clear" v-if="taskEditTask.labels.length" @mousedown.prevent.stop="clearAllLabels(props.search)"></div>
                    </template>
                </multiselect>
            </div>
        </div>

        <div class="field">
            <label class="label" for="subtasks">Subtasks</label>
            <div class="tasks noborder" v-if="taskEditTask.subtasks && taskEditTask.subtasks.length > 0">
                <div class="task" v-for="s in taskEditTask.subtasks" :key="s.id">
                    <label :for="s.id">
                        <div class="fancycheckbox">
                            <input @change="markAsDone" type="checkbox" :id="s.id" :checked="s.done" style="display: none;">
                            <label  :for="s.id" class="check">
                                <svg width="18px" height="18px" viewBox="0 0 18 18">
                                    <path d="M1,9 L1,3.5 C1,2 2,1 3.5,1 L14.5,1 C16,1 17,2 17,3.5 L17,14.5 C17,16 16,17 14.5,17 L3.5,17 C2,17 1,16 1,14.5 L1,9 Z"></path>
                                    <polyline points="1 9 7 14 15 4"></polyline>
                                </svg>
                            </label>
                        </div>
                        <span class="tasktext" :class="{ 'done': s.done}">
                                            {{s.text}}
                                        </span>
                    </label>
                </div>
            </div>
        </div>
        <div class="field has-addons">
            <div class="control is-expanded">
                <input @keyup.enter="addSubtask()" :class="{ 'disabled': taskService.loading}" :disabled="taskService.loading" class="input" type="text" id="tasktext" placeholder="New subtask" v-model="newTask.text"/>
            </div>
            <div class="control">
                <a class="button is-primary" @click="addSubtask()"><icon icon="plus"></icon></a>
            </div>
        </div>

        <button type="submit" class="button is-success is-fullwidth" :class="{ 'is-loading': taskService.loading}">
            Save
        </button>

    </form>
</template>

<script>
    import message from '../../message'
    import flatPickr from 'vue-flatpickr-component'
    import 'flatpickr/dist/flatpickr.css'
    import multiselect from 'vue-multiselect'
    import {differenceWith} from 'lodash'
	import verte from 'verte'
	import 'verte/dist/verte.css'

    import ListService from '../../services/list'
    import TaskService from '../../services/task'
    import TaskModel from '../../models/task'
    import UserModel from '../../models/user'
    import ListUserService from '../../services/listUsers'
    import priorities from '../../models/priorities'
    import LabelTaskService from '../../services/labelTask'
    import LabelService from '../../services/label'
    import LabelTaskModel from '../../models/labelTask'
    import LabelModel from '../../models/label'

    export default {
        name: 'edit-task',
        data() {
            return {
                listID: this.$route.params.id,
                listService: ListService,
                taskService: TaskService,

                priorities: priorities,
                list: {},
                newTask: TaskModel,
                isTaskEdit: false,
                taskEditTask: TaskModel,
                lastReminder: 0,
                nowUnix: new Date(),
                flatPickerConfig:{
                    altFormat: 'j M Y H:i',
                    altInput: true,
                    dateFormat: 'Y-m-d H:i',
                    enableTime: true,
                    onOpen: this.updateLastReminderDate,
                    onClose: this.addReminderDate,
                },

                newAssignee: UserModel,
                listUserService: ListUserService,
                foundUsers: [],

                labelService: LabelService,
                labelTaskService: LabelTaskService,
                foundLabels: [],
                labelTimeout: null,
            }
        },
        components: {
            flatPickr,
            multiselect,
            verte,
        },
        props: {
            task: {
                type: TaskModel,
                required: true,
            }
        },
        watch: {
            task() {
                this.taskEditTask = this.task
            }
        },
        created() {
            this.listService = new ListService()
            this.taskService = new TaskService()
            this.newTask = new TaskModel()
            this.listUserService = new ListUserService()
            this.newAssignee = new UserModel()
            this.labelService = new LabelService()
            this.labelTaskService = new LabelTaskService()
            this.taskEditTask = this.task
        },
        methods: {
            editTaskSubmit() {
                this.taskService.update(this.taskEditTask)
                    .then(r => {
                        this.$set(this, 'taskEditTask', r)
                        message.success({message: 'The task was successfully updated.'}, this)
                    })
                    .catch(e => {
                        message.error(e, this)
                    })
            },
            addSubtask() {
                this.newTask.parentTaskID = this.taskEditTask.id
                this.newTask.listID = this.$route.params.id
                this.taskService.create(this.newTask)
                    .then(r => {
                        this.list.addTaskToList(r)
                        message.success({message: 'The task was successfully created.'}, this)
                    })
                    .catch(e => {
                        message.error(e, this)
                    })

                this.newTask = {}
            },
            updateLastReminderDate(selectedDates) {
                this.lastReminder = +new Date(selectedDates[0])
            },
            addReminderDate(selectedDates, dateStr, instance) {
                let newDate = +new Date(selectedDates[0])

                // Don't update if nothing changed
                if (newDate === this.lastReminder) {
                    return
                }

                let index = parseInt(instance.input.dataset.index)
                this.taskEditTask.reminderDates[index] = newDate

                let lastIndex = this.taskEditTask.reminderDates.length - 1
                // put a new null at the end if we changed something
                if (lastIndex === index && !isNaN(newDate)) {
                    this.taskEditTask.reminderDates.push(null)
                }
            },
            removeReminderByIndex(index) {
                this.taskEditTask.reminderDates.splice(index, 1)
                // Reset the last to 0 to have the "add reminder" button
                this.taskEditTask.reminderDates[this.taskEditTask.reminderDates.length - 1] = null
            },
            addAssignee() {
                this.taskEditTask.assignees.push(this.newAssignee)
            },
            deleteAssigneeByIndex(index) {
                this.taskEditTask.assignees.splice(index, 1)
            },
            findUser(query) {
                if(query === '') {
                    this.clearAllFoundUsers()
                    return
                }

                this.listUserService.getAll({listID: this.$route.params.id}, {s: query})
                    .then(response => {
                        // Filter the results to not include users who are already assigned
                        this.$set(this, 'foundUsers', differenceWith(response, this.taskEditTask.assignees, (first, second) => {
                            return first.id === second.id
                        }))
                    })
                    .catch(e => {
                        message.error(e, this)
                    })
            },
            clearAllFoundUsers () {
                this.$set(this, 'foundUsers', [])
            },
            findLabel(query) {
                if(query === '') {
                    this.clearAllLabels()
                    return
                }

                if(this.labelTimeout !== null) {
                    clearTimeout(this.labelTimeout)
                }

                // Delay the search 300ms to not send a request on every keystroke
                this.labelTimeout = setTimeout(() => {
                    this.labelService.getAll({}, {s: query})
                        .then(response => {
                            this.$set(this, 'foundLabels', differenceWith(response, this.taskEditTask.labels, (first, second) => {
                                return first.id === second.id
                            }))
                            this.labelTimeout = null
                        })
                        .catch(e => {
                            message.error(e, this)
                        })
                }, 300)
            },
            clearAllLabels () {
                this.$set(this, 'foundLabels', [])
            },
            addLabel(label) {
                let labelTask = new LabelTaskModel({taskID: this.taskEditTask.id, label_id: label.id})
                this.labelTaskService.create(labelTask)
                    .then(() => {
                        message.success({message: 'The label was successfully added.'}, this)
                    })
                    .catch(e => {
                        message.error(e, this)
                    })
            },
            removeLabel(label) {
                let labelTask = new LabelTaskModel({taskID: this.taskEditTask.id, label_id: label.id})
                this.labelTaskService.delete(labelTask)
                    .then(() => {
                        // Remove the label from the list
                        for (const l in this.taskEditTask.labels) {
                            if (this.taskEditTask.labels[l].id === label.id) {
                                this.taskEditTask.labels.splice(l, 1)
                            }
                        }
                        message.success({message: 'The label was successfully removed.'}, this)
                    })
                    .catch(e => {
                        message.error(e, this)
                    })
            },
            createAndAddLabel(title) {
                let newLabel = new LabelModel({title: title})
                this.labelService.create(newLabel)
                    .then(r => {
                        this.addLabel(r)
                        this.taskEditTask.labels.push(r)
                    })
                    .catch(e => {
                        message.error(e, this)
                    })
            }
        },
    }
</script>

<style scoped>
	form {
		margin-bottom: 1em;
	}
</style>