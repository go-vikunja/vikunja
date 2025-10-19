import { describe, it, expect, beforeEach, vi } from 'vitest';
import { LabelResources, TeamResources, UserResources } from '../../../src/resources/other.js';
import { VikunjaClient } from '../../../src/vikunja/client.js';
import { UserContext } from '../../../src/auth/types.js';
import { VikunjaLabel, VikunjaTeam, VikunjaUser } from '../../../src/vikunja/types.js';

describe('Other Resources', () => {
  let mockClient: VikunjaClient;
  let mockUserContext: UserContext;

  beforeEach(() => {
    mockClient = {
      setToken: vi.fn(),
      get: vi.fn(),
      post: vi.fn(),
      put: vi.fn(),
      delete: vi.fn(),
    } as any;

    mockUserContext = {
      id: 1,
      username: 'testuser',
      email: 'test@example.com',
      token: 'test-token',
    };
  });

  describe('LabelResources', () => {
    let labelResources: LabelResources;

    beforeEach(() => {
      labelResources = new LabelResources(mockClient);
    });

    it('should list/read labels', async () => {
      const mockLabels: VikunjaLabel[] = [
        {
          id: 1,
          title: 'urgent',
          description: 'Urgent tasks',
          hex_color: '#ff0000',
          created: '2024-01-01',
          updated: '2024-01-01',
        },
      ];

      vi.spyOn(mockClient, 'get').mockResolvedValue(mockLabels);

      const result = await labelResources.list(mockUserContext);

      expect(mockClient.setToken).toHaveBeenCalledWith('test-token');
      expect(mockClient.get).toHaveBeenCalledWith('/api/v1/labels');
      expect(result).toHaveLength(1);
      expect(result[0].uri).toBe('vikunja://labels/1');
      expect(result[0].name).toBe('urgent');
    });

    it('should read single label', async () => {
      const mockLabel: VikunjaLabel = {
        id: 1,
        title: 'urgent',
        description: 'Urgent tasks',
        hex_color: '#ff0000',
        created: '2024-01-01',
        updated: '2024-01-01',
      };

      vi.spyOn(mockClient, 'get').mockResolvedValue(mockLabel);

      const result = await labelResources.read('vikunja://labels/1', mockUserContext);

      expect(mockClient.get).toHaveBeenCalledWith('/api/v1/labels/1');
      expect(result.contents[0].uri).toBe('vikunja://labels/1');
      expect(JSON.parse(result.contents[0].text)).toEqual(mockLabel);
    });
  });

  describe('TeamResources', () => {
    let teamResources: TeamResources;

    beforeEach(() => {
      teamResources = new TeamResources(mockClient);
    });

    it('should list/read teams', async () => {
      const mockTeams: VikunjaTeam[] = [
        {
          id: 1,
          name: 'Development Team',
          description: 'Dev team',
          members: [],
          created: '2024-01-01',
          updated: '2024-01-01',
        },
      ];

      vi.spyOn(mockClient, 'get').mockResolvedValue(mockTeams);

      const result = await teamResources.list(mockUserContext);

      expect(mockClient.setToken).toHaveBeenCalledWith('test-token');
      expect(mockClient.get).toHaveBeenCalledWith('/api/v1/teams');
      expect(result).toHaveLength(1);
      expect(result[0].uri).toBe('vikunja://teams/1');
      expect(result[0].name).toBe('Development Team');
    });

    it('should read single team', async () => {
      const mockTeam: VikunjaTeam = {
        id: 1,
        name: 'Development Team',
        description: 'Dev team',
        members: [],
        created: '2024-01-01',
        updated: '2024-01-01',
      };

      vi.spyOn(mockClient, 'get').mockResolvedValue(mockTeam);

      const result = await teamResources.read('vikunja://teams/1', mockUserContext);

      expect(mockClient.get).toHaveBeenCalledWith('/api/v1/teams/1');
      expect(JSON.parse(result.contents[0].text)).toEqual(mockTeam);
    });
  });

  describe('UserResources', () => {
    let userResources: UserResources;

    beforeEach(() => {
      userResources = new UserResources(mockClient);
    });

    it('should list/read users', async () => {
      const mockUsers: VikunjaUser[] = [
        {
          id: 1,
          username: 'testuser',
          email: 'test@example.com',
          name: 'Test User',
          created: '2024-01-01',
          updated: '2024-01-01',
        },
      ];

      vi.spyOn(mockClient, 'get').mockResolvedValue(mockUsers);

      const result = await userResources.list(mockUserContext, 'test');

      expect(mockClient.setToken).toHaveBeenCalledWith('test-token');
      expect(mockClient.get).toHaveBeenCalledWith('/api/v1/users', { s: 'test' });
      expect(result).toHaveLength(1);
      expect(result[0].uri).toBe('vikunja://users/1');
      expect(result[0].name).toBe('testuser');
    });

    it('should read current user', async () => {
      const mockUser: VikunjaUser = {
        id: 1,
        username: 'testuser',
        email: 'test@example.com',
        name: 'Test User',
        created: '2024-01-01',
        updated: '2024-01-01',
      };

      vi.spyOn(mockClient, 'get').mockResolvedValue(mockUser);

      const result = await userResources.read('vikunja://users/1', mockUserContext);

      expect(mockClient.get).toHaveBeenCalledWith('/api/v1/users/1');
      expect(JSON.parse(result.contents[0].text)).toEqual(mockUser);
    });

    it('should enforce user visibility permissions', async () => {
      vi.spyOn(mockClient, 'get').mockRejectedValue(new Error('Forbidden'));

      await expect(
        userResources.read('vikunja://users/999', mockUserContext)
      ).rejects.toThrow('Forbidden');
    });
  });
});
