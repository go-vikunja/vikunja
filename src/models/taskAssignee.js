import AbstractModel from './abstractModel'

export default class TaskAssigneeModel extends AbstractModel {
    constructor(data) {
        super(data)
        this.created = new Date(this.created)
    }

    defaults() {
        return {
            created: null,
            user_id: 0,
            task_id: 0,
        }
    }
}
