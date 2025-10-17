import { describe, it, expect, beforeEach, vi } from 'vitest';
import { VikunjaClient } from '../../../src/vikunja/client.js';
import { TeamAPI, UserAPI } from '../../../src/vikunja/users.js';
import { VikunjaTeam, VikunjaUser } from '../../../src/vikunja/types.js';

describe('Team/User API Methods', () => {
  let mockClient: VikunjaClient;
  let teamAPI: TeamAPI;
  let userAPI: UserAPI;

  const mockUser: VikunjaUser = {
    id: 1,
    username: 'user1',
    email: 'user1@example.com',
    name: 'User One',
    created: '2024-01-01T00:00:00Z',
    updated: '2024-01-01T00:00:00Z',
  };

  beforeEach(() => {
    mockClient = {
      get: vi.fn(),
      post: vi.fn(),
      put: vi.fn(),
      delete: vi.fn(),
      setToken: vi.fn(),
    } as unknown as VikunjaClient;

    teamAPI = new TeamAPI(mockClient);
    userAPI = new UserAPI(mockClient);
  });

  describe('TeamAPI', () => {
    describe('getTeams', () => {
      it('should get all teams', async () => {
        const mockTeams: VikunjaTeam[] = [
          {
            id: 1,
            name: 'Development Team',
            description: 'Dev team',
            members: [mockUser],
            created: '2024-01-01T00:00:00Z',
            updated: '2024-01-01T00:00:00Z',
          },
          {
            id: 2,
            name: 'QA Team',
            description: 'Quality assurance',
            members: [],
            created: '2024-01-02T00:00:00Z',
            updated: '2024-01-02T00:00:00Z',
          },
        ];

        vi.mocked(mockClient.get).mockResolvedValue(mockTeams);

        const teams = await teamAPI.getTeams();
        expect(teams).toEqual(mockTeams);
        expect(mockClient.get).toHaveBeenCalledWith('/api/v1/teams');
      });
    });

    describe('getTeam', () => {
      it('should get single team', async () => {
        const mockTeam: VikunjaTeam = {
          id: 1,
          name: 'Development Team',
          description: 'Dev team',
          members: [mockUser],
          created: '2024-01-01T00:00:00Z',
          updated: '2024-01-01T00:00:00Z',
        };

        vi.mocked(mockClient.get).mockResolvedValue(mockTeam);

        const team = await teamAPI.getTeam(1);
        expect(team).toEqual(mockTeam);
        expect(mockClient.get).toHaveBeenCalledWith('/api/v1/teams/1');
      });
    });
  });

  describe('UserAPI', () => {
    describe('getCurrentUser', () => {
      it('should get current user', async () => {
        vi.mocked(mockClient.get).mockResolvedValue(mockUser);

        const user = await userAPI.getCurrentUser();
        expect(user).toEqual(mockUser);
        expect(mockClient.get).toHaveBeenCalledWith('/api/v1/user');
      });
    });

    describe('searchUsers', () => {
      it('should search users', async () => {
        const mockUsers: VikunjaUser[] = [
          mockUser,
          {
            id: 2,
            username: 'user2',
            email: 'user2@example.com',
            name: 'User Two',
            created: '2024-01-02T00:00:00Z',
            updated: '2024-01-02T00:00:00Z',
          },
        ];

        vi.mocked(mockClient.get).mockResolvedValue(mockUsers);

        const users = await userAPI.searchUsers('user');
        expect(users).toEqual(mockUsers);
        expect(mockClient.get).toHaveBeenCalledWith('/api/v1/users', { s: 'user' });
      });
    });
  });
});
