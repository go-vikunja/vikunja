import AbstractModel from './abstractModel'
import UserModel from './user'

export default class ListModel extends AbstractModel {

    constructor(data) {
        // The constructor of AbstractModel handles all the default parsing.
        super(data)

        this.shared_by = new UserModel(this.shared_by)
    }

    // Default attributes that define the "empty" state.
    defaults() {
        return {
            id: 0,
            hash: '',
            right: 0,
            shared_by: UserModel,
            sharing_type: 0,
            listID: 0,

            created: 0,
            updated: 0,
        }
    }
}