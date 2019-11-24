import AbstractModel from './abstractModel'

export default class TaskAssigneeModel extends AbstractModel {
    defaults() {
        return {
            created: 0,
            user_id: 0,
            task_id: 0,
        }
    }
}
