import {Factory} from '../support/factory'
import {formatISO} from 'date-fns'

export class TeamMemberFactory extends Factory {
    static table = 'team_members'

    static factory() {
        return {
            team_id: 1,
            user_id: 1,
            admin: false,
            created: formatISO(new Date()),
        }
    }
}