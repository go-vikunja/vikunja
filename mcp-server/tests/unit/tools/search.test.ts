import { describe, it, expect, beforeEach, vi } from 'vitest';
import {
  SearchTools,
  SearchTasksSchema,
  GetMyTasksSchema,
} from '../../../src/tools/search.js';
import { VikunjaClient } from '../../../src/vikunja/client.js';
import { RateLimiter } from '../../../src/ratelimit/limiter.js';
import { UserContext } from '../../../src/auth/types.js';
import { VikunjaTask, VikunjaProject } from '../../../src/vikunja/types.js';

describe('Search Tools', () => {
  let searchTools: SearchTools;
  let mockClient: VikunjaClient;
  let mockRateLimiter: RateLimiter;
  let userContext: UserContext;

  const mockTask: VikunjaTask = {
    id: 1,
    title: 'Test Task',
    description: '',
    done: false,
    done_at: null,
    due_date: null,
    priority: 3,
    labels: [],
    assignees: [],
    project_id: 1,
    created: '2025-01-01T00:00:00Z',
    updated: '2025-01-01T00:00:00Z',
    created_by: {
      id: 1,
      username: 'testuser',
      email: 'test@example.com',
      name: 'Test User',
      created: '2025-01-01T00:00:00Z',
      updated: '2025-01-01T00:00:00Z',
    },
  };

  const mockProject: VikunjaProject = {
    id: 1,
    title: 'Test Project',
    description: '',
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

    searchTools = new SearchTools(mockClient, mockRateLimiter);

    userContext = {
      userId: 1,
      username: 'testuser',
      email: 'test@example.com',
      token: 'test-token',
    };
  });

  describe('searchTasks', () => {
    it('should search tasks by query', async () => {
      const input = {
        query: 'test',
        page: 1,
      };

      vi.mocked(mockClient.get).mockResolvedValue([mockTask]);

      const result = await searchTools.searchTasks(input, userContext);

      expect(result.success).toBe(true);
      expect(result.tasks).toHaveLength(1);
      expect(result.total).toBe(1);
      expect(mockClient.get).toHaveBeenCalledWith('/api/v1/tasks/all', {
        s: 'test',
        page: 1,
      });
    });

    it('should apply label filters', async () => {
      const input = {
        query: 'test',
        page: 1,
        filter_labels: [1, 2],
      };

      const taskWithLabels = {
        ...mockTask,
        labels: [{ id: 1, title: 'Label 1', description: '', hex_color: '', created: '', updated: '' }],
      };

      vi.mocked(mockClient.get).mockResolvedValue([taskWithLabels]);

      const result = await searchTools.searchTasks(input, userContext);

      expect(result.success).toBe(true);
      expect(result.tasks).toHaveLength(1);
    });

    it('should validate input with Zod schema', () => {
      const invalidInput = {
        query: '', // Empty query should fail
      };

      const result = SearchTasksSchema.safeParse(invalidInput);
      expect(result.success).toBe(false);
    });
  });

  describe('searchProjects', () => {
    it('should search projects by query', async () => {
      const input = {
        query: 'test',
        page: 1,
      };

      vi.mocked(mockClient.get).mockResolvedValue([mockProject]);

      const result = await searchTools.searchProjects(input, userContext);

      expect(result.success).toBe(true);
      expect(result.projects).toHaveLength(1);
      expect(mockClient.get).toHaveBeenCalledWith('/api/v1/projects', {
        s: 'test',
        page: 1,
      });
    });

    it('should apply archive filter', async () => {
      const input = {
        query: 'test',
        page: 1,
        filter_archived: true,
      };

      vi.mocked(mockClient.get).mockResolvedValue([mockProject]);

      const result = await searchTools.searchProjects(input, userContext);

      expect(result.success).toBe(true);
      expect(mockClient.get).toHaveBeenCalledWith(
        '/api/v1/projects',
        expect.objectContaining({
          is_archived: true,
        })
      );
    });
  });

  describe('getMyTasks', () => {
    it('should get current user tasks', async () => {
      const input = {
        page: 1,
      };

      vi.mocked(mockClient.get).mockResolvedValue([mockTask]);

      const result = await searchTools.getMyTasks(input, userContext);

      expect(result.success).toBe(true);
      expect(result.tasks).toHaveLength(1);
      expect(mockClient.get).toHaveBeenCalledWith(
        '/api/v1/tasks/all',
        expect.objectContaining({
          filter_by: 'assignees',
          filter_value: 1,
        })
      );
    });

    it('should validate input with Zod schema', () => {
      const validInput = {
        page: 1,
        filter_done: false,
      };

      const result = GetMyTasksSchema.safeParse(validInput);
      expect(result.success).toBe(true);
    });
  });

  describe('getProjectTasks', () => {
    it('should get all tasks in a project', async () => {
      const input = {
        project_id: 1,
        page: 1,
      };

      vi.mocked(mockClient.get).mockResolvedValue([mockTask]);

      const result = await searchTools.getProjectTasks(input, userContext);

      expect(result.success).toBe(true);
      expect(result.tasks).toHaveLength(1);
      expect(mockClient.get).toHaveBeenCalledWith('/api/v1/projects/1/tasks', {
        page: 1,
      });
    });

    it('should apply priority filter', async () => {
      const input = {
        project_id: 1,
        page: 1,
        filter_priority: 5,
      };

      vi.mocked(mockClient.get).mockResolvedValue([mockTask]);

      const result = await searchTools.getProjectTasks(input, userContext);

      expect(result.success).toBe(true);
      expect(mockClient.get).toHaveBeenCalledWith(
        '/api/v1/projects/1/tasks',
        expect.objectContaining({
          filter_by: 'priority',
          filter_value: 5,
        })
      );
    });
  });
});
