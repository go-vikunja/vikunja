import { describe, it, expect, beforeEach, vi } from 'vitest';
import {
  BulkTools,
  BulkUpdateTasksSchema,
  BulkCompleteTasksSchema,
} from '../../../src/tools/bulk.js';
import { VikunjaClient } from '../../../src/vikunja/client.js';
import { RateLimiter } from '../../../src/ratelimit/limiter.js';
import { UserContext } from '../../../src/auth/types.js';
import { VikunjaTask } from '../../../src/vikunja/types.js';

describe('Bulk Tools', () => {
  let bulkTools: BulkTools;
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
    priority: 0,
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

    bulkTools = new BulkTools(mockClient, mockRateLimiter);

    userContext = {
      userId: 1,
      username: 'testuser',
      email: 'test@example.com',
      token: 'test-token',
    };
  });

  describe('bulkUpdateTasks', () => {
    it('should update multiple tasks successfully', async () => {
      const input = {
        task_ids: [1, 2, 3],
        update_data: {
          priority: 5,
        },
      };

      vi.mocked(mockClient.post).mockResolvedValue(mockTask);

      const result = await bulkTools.bulkUpdateTasks(input, userContext);

      expect(result.success).toBe(true);
      expect(result.successCount).toBe(3);
      expect(result.failureCount).toBe(0);
      expect(mockClient.post).toHaveBeenCalledTimes(3);
    });

    it('should handle partial failures', async () => {
      const input = {
        task_ids: [1, 2, 3],
        update_data: {
          priority: 5,
        },
      };

      vi.mocked(mockClient.post)
        .mockResolvedValueOnce(mockTask)
        .mockRejectedValueOnce(new Error('Failed'))
        .mockResolvedValueOnce(mockTask);

      const result = await bulkTools.bulkUpdateTasks(input, userContext);

      expect(result.success).toBe(false);
      expect(result.successCount).toBe(2);
      expect(result.failureCount).toBe(1);
      expect(result.errors).toHaveLength(1);
      expect(result.errors?.[0].taskId).toBe(2);
    });

    it('should enforce batch size limit', () => {
      const invalidInput = {
        task_ids: Array(101).fill(1), // 101 tasks exceeds limit
        update_data: {},
      };

      const result = BulkUpdateTasksSchema.safeParse(invalidInput);
      expect(result.success).toBe(false);
    });

    it('should validate input with Zod schema', () => {
      const validInput = {
        task_ids: [1, 2],
        update_data: {
          priority: 3,
        },
      };

      const result = BulkUpdateTasksSchema.safeParse(validInput);
      expect(result.success).toBe(true);
    });
  });

  describe('bulkCompleteTasks', () => {
    it('should complete multiple tasks', async () => {
      const input = {
        task_ids: [1, 2],
      };

      vi.mocked(mockClient.post).mockResolvedValue({
        ...mockTask,
        done: true,
      });

      const result = await bulkTools.bulkCompleteTasks(input, userContext);

      expect(result.success).toBe(true);
      expect(result.successCount).toBe(2);
      expect(mockClient.post).toHaveBeenCalledWith('/api/v1/tasks/1', { done: true });
    });

    it('should validate input with Zod schema', () => {
      const invalidInput = {
        task_ids: [], // Empty array should fail
      };

      const result = BulkCompleteTasksSchema.safeParse(invalidInput);
      expect(result.success).toBe(false);
    });
  });

  describe('bulkAssignTasks', () => {
    it('should assign user to multiple tasks', async () => {
      const input = {
        task_ids: [1, 2],
        user_id: 5,
      };

      vi.mocked(mockClient.put).mockResolvedValue(mockTask);

      const result = await bulkTools.bulkAssignTasks(input, userContext);

      expect(result.success).toBe(true);
      expect(result.successCount).toBe(2);
      expect(mockClient.put).toHaveBeenCalledWith('/api/v1/tasks/1/assignees', {
        user_id: 5,
      });
    });
  });

  describe('bulkAddLabels', () => {
    it('should add label to multiple tasks', async () => {
      const input = {
        task_ids: [1, 2],
        label_id: 3,
      };

      vi.mocked(mockClient.put).mockResolvedValue(mockTask);

      const result = await bulkTools.bulkAddLabels(input, userContext);

      expect(result.success).toBe(true);
      expect(result.successCount).toBe(2);
      expect(mockClient.put).toHaveBeenCalledWith('/api/v1/tasks/1/labels', {
        label_id: 3,
      });
    });
  });
});
