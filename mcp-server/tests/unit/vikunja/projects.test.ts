import { describe, it, expect, beforeEach, vi } from 'vitest';
import { VikunjaClient } from '../../../src/vikunja/client.js';
import { ProjectAPI } from '../../../src/vikunja/projects.js';
import { VikunjaProject } from '../../../src/vikunja/types.js';
import { PermissionError } from '../../../src/utils/errors.js';

describe('Project API Methods', () => {
  let mockClient: VikunjaClient;
  let projectAPI: ProjectAPI;

  beforeEach(() => {
    mockClient = {
      get: vi.fn(),
      post: vi.fn(),
      put: vi.fn(),
      delete: vi.fn(),
      setToken: vi.fn(),
    } as unknown as VikunjaClient;

    projectAPI = new ProjectAPI(mockClient);
  });

  describe('getProjects', () => {
    it('should get all projects', async () => {
      const mockProjects: VikunjaProject[] = [
        {
          id: 1,
          title: 'Project 1',
          description: 'Test project 1',
          owner: {
            id: 1,
            username: 'user1',
            email: 'user1@example.com',
            name: 'User One',
            created: '2024-01-01T00:00:00Z',
            updated: '2024-01-01T00:00:00Z',
          },
          created: '2024-01-01T00:00:00Z',
          updated: '2024-01-01T00:00:00Z',
          is_archived: false,
          hex_color: '#ff0000',
          parent_project_id: 0,
        },
        {
          id: 2,
          title: 'Project 2',
          description: 'Test project 2',
          owner: {
            id: 1,
            username: 'user1',
            email: 'user1@example.com',
            name: 'User One',
            created: '2024-01-01T00:00:00Z',
            updated: '2024-01-01T00:00:00Z',
          },
          created: '2024-01-02T00:00:00Z',
          updated: '2024-01-02T00:00:00Z',
          is_archived: false,
          hex_color: '#00ff00',
          parent_project_id: 0,
        },
      ];

      vi.mocked(mockClient.get).mockResolvedValue(mockProjects);

      const projects = await projectAPI.getProjects();
      expect(projects).toEqual(mockProjects);
      expect(mockClient.get).toHaveBeenCalledWith('/api/v1/projects', { page: 1 });
    });

    it('should handle pagination', async () => {
      const mockProjects: VikunjaProject[] = [
        {
          id: 51,
          title: 'Project 51',
          description: 'Test project on page 2',
          owner: {
            id: 1,
            username: 'user1',
            email: 'user1@example.com',
            name: 'User One',
            created: '2024-01-01T00:00:00Z',
            updated: '2024-01-01T00:00:00Z',
          },
          created: '2024-01-01T00:00:00Z',
          updated: '2024-01-01T00:00:00Z',
          is_archived: false,
          hex_color: '#0000ff',
          parent_project_id: 0,
        },
      ];

      vi.mocked(mockClient.get).mockResolvedValue(mockProjects);

      const projects = await projectAPI.getProjects(2);
      expect(projects).toEqual(mockProjects);
      expect(mockClient.get).toHaveBeenCalledWith('/api/v1/projects', { page: 2 });
    });
  });

  describe('getProject', () => {
    it('should get single project', async () => {
      const mockProject: VikunjaProject = {
        id: 1,
        title: 'Project 1',
        description: 'Test project 1',
        owner: {
          id: 1,
          username: 'user1',
          email: 'user1@example.com',
          name: 'User One',
          created: '2024-01-01T00:00:00Z',
          updated: '2024-01-01T00:00:00Z',
        },
        created: '2024-01-01T00:00:00Z',
        updated: '2024-01-01T00:00:00Z',
        is_archived: false,
        hex_color: '#ff0000',
        parent_project_id: 0,
      };

      vi.mocked(mockClient.get).mockResolvedValue(mockProject);

      const project = await projectAPI.getProject(1);
      expect(project).toEqual(mockProject);
      expect(mockClient.get).toHaveBeenCalledWith('/api/v1/projects/1');
    });
  });

  describe('createProject', () => {
    it('should create project', async () => {
      const input = {
        title: 'New Project',
        description: 'A new test project',
        hex_color: '#ff0000',
      };

      const mockProject: VikunjaProject = {
        id: 3,
        title: input.title,
        description: input.description,
        owner: {
          id: 1,
          username: 'user1',
          email: 'user1@example.com',
          name: 'User One',
          created: '2024-01-01T00:00:00Z',
          updated: '2024-01-01T00:00:00Z',
        },
        created: '2024-01-03T00:00:00Z',
        updated: '2024-01-03T00:00:00Z',
        is_archived: false,
        hex_color: input.hex_color,
        parent_project_id: 0,
      };

      vi.mocked(mockClient.post).mockResolvedValue(mockProject);

      const project = await projectAPI.createProject(input);
      expect(project).toEqual(mockProject);
      expect(mockClient.post).toHaveBeenCalledWith('/api/v1/projects', input);
    });
  });

  describe('updateProject', () => {
    it('should update project', async () => {
      const input = {
        title: 'Updated Project',
        description: 'Updated description',
      };

      const mockProject: VikunjaProject = {
        id: 1,
        title: input.title,
        description: input.description,
        owner: {
          id: 1,
          username: 'user1',
          email: 'user1@example.com',
          name: 'User One',
          created: '2024-01-01T00:00:00Z',
          updated: '2024-01-01T00:00:00Z',
        },
        created: '2024-01-01T00:00:00Z',
        updated: '2024-01-03T00:00:00Z',
        is_archived: false,
        hex_color: '#ff0000',
        parent_project_id: 0,
      };

      vi.mocked(mockClient.put).mockResolvedValue(mockProject);

      const project = await projectAPI.updateProject(1, input);
      expect(project).toEqual(mockProject);
      expect(mockClient.put).toHaveBeenCalledWith('/api/v1/projects/1', input);
    });
  });

  describe('deleteProject', () => {
    it('should delete project', async () => {
      vi.mocked(mockClient.delete).mockResolvedValue({ message: 'Successfully deleted.' });

      await projectAPI.deleteProject(1);
      expect(mockClient.delete).toHaveBeenCalledWith('/api/v1/projects/1');
    });
  });

  describe('error handling', () => {
    it('should map 403 to PermissionError', async () => {
      vi.mocked(mockClient.get).mockRejectedValue(
        new PermissionError('You do not have permission to access this project')
      );

      await expect(projectAPI.getProject(1)).rejects.toThrow(PermissionError);
    });
  });
});
