import AbstractModel from './abstractModel'

export default class TaskAssigneeModel extends AbstractModel {
    constructor(data) {
        super(data)
        this.created = new Date(this.created)
    }

    defaults() {
        return {
            created: null,
            userId: 0,
            taskId: 0,
        }
    }
}
