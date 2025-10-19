import { describe, it, expect, beforeEach, vi } from 'vitest';
import { ProjectTools, CreateProjectSchema, UpdateProjectSchema } from '../../../src/tools/projects.js';
import { VikunjaClient } from '../../../src/vikunja/client.js';
import { RateLimiter } from '../../../src/ratelimit/limiter.js';
import { UserContext } from '../../../src/auth/types.js';
import { VikunjaProject } from '../../../src/vikunja/types.js';

describe('Project Tools', () => {
  let projectTools: ProjectTools;
  let mockClient: VikunjaClient;
  let mockRateLimiter: RateLimiter;
  let userContext: UserContext;

  const mockProject: VikunjaProject = {
    id: 1,
    title: 'Test Project',
    description: 'Test Description',
    owner: {
      id: 1,
      username: 'testuser',
      email: 'test@example.com',
      name: 'Test User',
      created: '2025-01-01T00:00:00Z',
      updated: '2025-01-01T00:00:00Z',
    },
    created: '2025-01-01T00:00:00Z',
    updated: '2025-01-01T00:00:00Z',
    is_archived: false,
    hex_color: '#ffffff',
    parent_project_id: 0,
  };

  beforeEach(() => {
    mockClient = {
      setToken: vi.fn(),
      get: vi.fn(),
      post: vi.fn(),
      put: vi.fn(),
      delete: vi.fn(),
    } as unknown as VikunjaClient;

    mockRateLimiter = {
      checkLimit: vi.fn().mockResolvedValue(undefined),
    } as unknown as RateLimiter;

    projectTools = new ProjectTools(mockClient, mockRateLimiter);

    userContext = {
      userId: 1,
      username: 'testuser',
      email: 'test@example.com',
      token: 'test-token',
    };
  });

  describe('createProject', () => {
    it('should create a project successfully', async () => {
      const input = {
        title: 'New Project',
        description: 'Project description',
      };

      vi.mocked(mockClient.post).mockResolvedValue(mockProject);

      const result = await projectTools.createProject(input, userContext);

      expect(result.success).toBe(true);
      expect(result.message).toContain('created successfully');
      expect(result.project).toEqual(mockProject);
      expect(mockRateLimiter.checkLimit).toHaveBeenCalledWith('test-token');
      expect(mockClient.setToken).toHaveBeenCalledWith('test-token');
      expect(mockClient.post).toHaveBeenCalledWith('/api/v1/projects', input);
    });

    it('should validate input with Zod schema', () => {
      const invalidInput = {
        title: '', // Empty title should fail
      };

      const result = CreateProjectSchema.safeParse(invalidInput);
      expect(result.success).toBe(false);
    });

    it('should handle Vikunja API errors', async () => {
      const input = {
        title: 'New Project',
      };

      vi.mocked(mockClient.post).mockRejectedValue(new Error('API Error'));

      const result = await projectTools.createProject(input, userContext);

      expect(result.success).toBe(false);
      expect(result.message).toBe('Failed to create project');
      expect(result.error).toBe('API Error');
    });

    it('should enforce rate limits', async () => {
      const input = {
        title: 'New Project',
      };

      vi.mocked(mockRateLimiter.checkLimit).mockRejectedValue(
        new Error('Rate limit exceeded')
      );

      const result = await projectTools.createProject(input, userContext);

      expect(result.success).toBe(false);
      expect(result.error).toBe('Rate limit exceeded');
    });
  });

  describe('updateProject', () => {
    it('should update a project successfully', async () => {
      const input = {
        id: 1,
        title: 'Updated Project',
      };

      vi.mocked(mockClient.put).mockResolvedValue(mockProject);

      const result = await projectTools.updateProject(input, userContext);

      expect(result.success).toBe(true);
      expect(result.message).toContain('updated successfully');
      expect(result.project).toEqual(mockProject);
      expect(mockClient.put).toHaveBeenCalledWith('/api/v1/projects/1', {
        title: 'Updated Project',
      });
    });

    it('should validate input with Zod schema', () => {
      const invalidInput = {
        id: -1, // Negative ID should fail
        title: 'Updated Project',
      };

      const result = UpdateProjectSchema.safeParse(invalidInput);
      expect(result.success).toBe(false);
    });

    it('should handle update errors', async () => {
      const input = {
        id: 1,
        title: 'Updated Project',
      };

      vi.mocked(mockClient.put).mockRejectedValue(new Error('Not found'));

      const result = await projectTools.updateProject(input, userContext);

      expect(result.success).toBe(false);
      expect(result.error).toBe('Not found');
    });
  });

  describe('deleteProject', () => {
    it('should delete a project successfully', async () => {
      const input = {
        id: 1,
      };

      vi.mocked(mockClient.delete).mockResolvedValue(undefined);

      const result = await projectTools.deleteProject(input, userContext);

      expect(result.success).toBe(true);
      expect(result.message).toContain('deleted successfully');
      expect(mockClient.delete).toHaveBeenCalledWith('/api/v1/projects/1');
    });

    it('should handle delete errors', async () => {
      const input = {
        id: 1,
      };

      vi.mocked(mockClient.delete).mockRejectedValue(new Error('Permission denied'));

      const result = await projectTools.deleteProject(input, userContext);

      expect(result.success).toBe(false);
      expect(result.error).toBe('Permission denied');
    });
  });

  describe('archiveProject', () => {
    it('should archive a project successfully', async () => {
      const input = {
        id: 1,
        archived: true,
      };

      vi.mocked(mockClient.put).mockResolvedValue({
        ...mockProject,
        is_archived: true,
      });

      const result = await projectTools.archiveProject(input, userContext);

      expect(result.success).toBe(true);
      expect(result.message).toContain('archived successfully');
      expect(mockClient.put).toHaveBeenCalledWith('/api/v1/projects/1', {
        is_archived: true,
      });
    });

    it('should unarchive a project successfully', async () => {
      const input = {
        id: 1,
        archived: false,
      };

      vi.mocked(mockClient.put).mockResolvedValue(mockProject);

      const result = await projectTools.archiveProject(input, userContext);

      expect(result.success).toBe(true);
      expect(result.message).toContain('unarchived successfully');
    });

    it('should handle archive errors', async () => {
      const input = {
        id: 1,
        archived: true,
      };

      vi.mocked(mockClient.put).mockRejectedValue(new Error('Archive failed'));

      const result = await projectTools.archiveProject(input, userContext);

      expect(result.success).toBe(false);
      expect(result.error).toBe('Archive failed');
    });
  });
});
