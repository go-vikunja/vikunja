import {Factory} from '../support/factory'

export class TeamMemberFactory extends Factory {
    static table = 'team_members'

    static factory() {
        return {
            team_id: 1,
            user_id: 1,
            admin: false,
            created: new Date().toISOString(),
        }
    }
}
