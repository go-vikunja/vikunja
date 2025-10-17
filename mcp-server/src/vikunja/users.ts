import { VikunjaClient } from './client.js';
import { VikunjaTeam, VikunjaUser } from './types.js';

/**
 * Team API methods
 */
export class TeamAPI {
  constructor(private client: VikunjaClient) {}

  /**
   * Get all teams
   * @returns List of teams
   */
  async getTeams(): Promise<VikunjaTeam[]> {
    return this.client.get<VikunjaTeam[]>('/api/v1/teams');
  }

  /**
   * Get a single team by ID
   * @param id - Team ID
   * @returns Team details
   */
  async getTeam(id: number): Promise<VikunjaTeam> {
    return this.client.get<VikunjaTeam>(`/api/v1/teams/${id}`);
  }
}

/**
 * User API methods
 */
export class UserAPI {
  constructor(private client: VikunjaClient) {}

  /**
   * Get current user
   * @returns Current user details
   */
  async getCurrentUser(): Promise<VikunjaUser> {
    return this.client.get<VikunjaUser>('/api/v1/user');
  }

  /**
   * Search for users
   * @param query - Search query
   * @returns List of matching users
   */
  async searchUsers(query: string): Promise<VikunjaUser[]> {
    return this.client.get<VikunjaUser[]>('/api/v1/users', { s: query });
  }
}
