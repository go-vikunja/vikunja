import { describe, it, expect, beforeEach, vi } from 'vitest';
import { ProjectResources } from '../../../src/resources/projects.js';
import { VikunjaClient } from '../../../src/vikunja/client.js';
import { UserContext } from '../../../src/auth/types.js';
import { VikunjaProject } from '../../../src/vikunja/types.js';

describe('Project Resources', () => {
  let projectResources: ProjectResources;
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

    projectResources = new ProjectResources(mockClient);
  });

  describe('list', () => {
    it('should list all projects', async () => {
      const mockProjects: VikunjaProject[] = [
        {
          id: 1,
          title: 'Project 1',
          description: 'Description 1',
          owner: { id: 1, username: 'testuser', email: 'test@example.com', name: 'Test User', created: '2024-01-01', updated: '2024-01-01' },
          created: '2024-01-01',
          updated: '2024-01-01',
          is_archived: false,
          hex_color: '#ffffff',
          parent_project_id: 0,
        },
        {
          id: 2,
          title: 'Project 2',
          description: 'Description 2',
          owner: { id: 1, username: 'testuser', email: 'test@example.com', name: 'Test User', created: '2024-01-01', updated: '2024-01-01' },
          created: '2024-01-01',
          updated: '2024-01-01',
          is_archived: false,
          hex_color: '#000000',
          parent_project_id: 0,
        },
      ];

      vi.spyOn(mockClient, 'get').mockResolvedValue(mockProjects);

      const result = await projectResources.list(mockUserContext);

      expect(mockClient.setToken).toHaveBeenCalledWith('test-token');
      expect(mockClient.get).toHaveBeenCalledWith('/api/v1/projects', { page: 1 });
      expect(result).toHaveLength(2);
      expect(result[0].uri).toBe('vikunja://projects/1');
      expect(result[0].name).toBe('Project 1');
      expect(result[0].mimeType).toBe('application/json');
    });

    it('should handle pagination', async () => {
      const mockProjects: VikunjaProject[] = [];
      vi.spyOn(mockClient, 'get').mockResolvedValue(mockProjects);

      await projectResources.list(mockUserContext, 2);

      expect(mockClient.get).toHaveBeenCalledWith('/api/v1/projects', { page: 2 });
    });
  });

  describe('read', () => {
    it('should read single project', async () => {
      const mockProject: VikunjaProject = {
        id: 1,
        title: 'Test Project',
        description: 'Test Description',
        owner: { id: 1, username: 'testuser', email: 'test@example.com', name: 'Test User', created: '2024-01-01', updated: '2024-01-01' },
        created: '2024-01-01',
        updated: '2024-01-01',
        is_archived: false,
        hex_color: '#ffffff',
        parent_project_id: 0,
      };

      vi.spyOn(mockClient, 'get').mockResolvedValue(mockProject);

      const result = await projectResources.read('vikunja://projects/1', mockUserContext);

      expect(mockClient.setToken).toHaveBeenCalledWith('test-token');
      expect(mockClient.get).toHaveBeenCalledWith('/api/v1/projects/1');
      expect(result.contents).toHaveLength(1);
      expect(result.contents[0].uri).toBe('vikunja://projects/1');
      expect(result.contents[0].mimeType).toBe('application/json');
      expect(JSON.parse(result.contents[0].text)).toEqual(mockProject);
    });

    it('should include metadata', async () => {
      const mockProject: VikunjaProject = {
        id: 1,
        title: 'Test Project',
        description: 'Test Description',
        owner: { id: 1, username: 'testuser', email: 'test@example.com', name: 'Test User', created: '2024-01-01', updated: '2024-01-01' },
        created: '2024-01-01',
        updated: '2024-01-01',
        is_archived: false,
        hex_color: '#ffffff',
        parent_project_id: 0,
      };

      vi.spyOn(mockClient, 'get').mockResolvedValue(mockProject);

      const result = await projectResources.read('vikunja://projects/1', mockUserContext);

      const parsed = JSON.parse(result.contents[0].text);
      expect(parsed.id).toBe(1);
      expect(parsed.title).toBe('Test Project');
      expect(parsed.owner.username).toBe('testuser');
    });

    it('should handle not found', async () => {
      vi.spyOn(mockClient, 'get').mockRejectedValue(new Error('Project not found'));

      await expect(
        projectResources.read('vikunja://projects/999', mockUserContext)
      ).rejects.toThrow('Project not found');
    });

    it('should parse project ID from URI', async () => {
      const mockProject: VikunjaProject = {
        id: 42,
        title: 'Test',
        description: '',
        owner: { id: 1, username: 'testuser', email: 'test@example.com', name: 'Test User', created: '2024-01-01', updated: '2024-01-01' },
        created: '2024-01-01',
        updated: '2024-01-01',
        is_archived: false,
        hex_color: '#ffffff',
        parent_project_id: 0,
      };

      vi.spyOn(mockClient, 'get').mockResolvedValue(mockProject);

      await projectResources.read('vikunja://projects/42', mockUserContext);

      expect(mockClient.get).toHaveBeenCalledWith('/api/v1/projects/42');
    });
  });
});
