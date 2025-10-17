import { describe, it, expect, beforeEach, vi } from 'vitest';
import {
  TaskTools,
  CreateTaskSchema,
  UpdateTaskSchema,
  MoveTaskSchema,
} from '../../../src/tools/tasks.js';
import { VikunjaClient } from '../../../src/vikunja/client.js';
import { RateLimiter } from '../../../src/ratelimit/limiter.js';
import { UserContext } from '../../../src/auth/types.js';
import { VikunjaTask } from '../../../src/vikunja/types.js';

describe('Task Tools', () => {
  let taskTools: TaskTools;
  let mockClient: VikunjaClient;
  let mockRateLimiter: RateLimiter;
  let userContext: UserContext;

  const mockTask: VikunjaTask = {
    id: 1,
    title: 'Test Task',
    description: 'Test Description',
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

    taskTools = new TaskTools(mockClient, mockRateLimiter);

    userContext = {
      userId: 1,
      username: 'testuser',
      email: 'test@example.com',
      token: 'test-token',
    };
  });

  describe('createTask', () => {
    it('should create a task successfully', async () => {
      const input = {
        project_id: 1,
        title: 'New Task',
        description: 'Task description',
      };

      vi.mocked(mockClient.put).mockResolvedValue(mockTask);

      const result = await taskTools.createTask(input, userContext);

      expect(result.success).toBe(true);
      expect(result.message).toContain('created successfully');
      expect(result.task).toEqual(mockTask);
      expect(result.taskId).toBe(1);
      expect(mockClient.put).toHaveBeenCalledWith('/api/v1/projects/1', input);
    });

    it('should validate input with Zod schema', () => {
      const invalidInput = {
        project_id: -1, // Negative ID should fail
        title: 'New Task',
      };

      const result = CreateTaskSchema.safeParse(invalidInput);
      expect(result.success).toBe(false);
    });

    it('should handle creation errors', async () => {
      const input = {
        project_id: 1,
        title: 'New Task',
      };

      vi.mocked(mockClient.put).mockRejectedValue(new Error('Creation failed'));

      const result = await taskTools.createTask(input, userContext);

      expect(result.success).toBe(false);
      expect(result.error).toBe('Creation failed');
    });
  });

  describe('updateTask', () => {
    it('should update a task successfully', async () => {
      const input = {
        id: 1,
        title: 'Updated Task',
        priority: 5,
      };

      vi.mocked(mockClient.post).mockResolvedValue({
        ...mockTask,
        title: 'Updated Task',
        priority: 5,
      });

      const result = await taskTools.updateTask(input, userContext);

      expect(result.success).toBe(true);
      expect(result.message).toContain('updated successfully');
      expect(mockClient.post).toHaveBeenCalledWith('/api/v1/tasks/1', {
        title: 'Updated Task',
        priority: 5,
      });
    });

    it('should validate input with Zod schema', () => {
      const invalidInput = {
        id: 1,
        priority: 10, // Priority > 5 should fail
      };

      const result = UpdateTaskSchema.safeParse(invalidInput);
      expect(result.success).toBe(false);
    });
  });

  describe('completeTask', () => {
    it('should mark a task as complete', async () => {
      const input = {
        id: 1,
      };

      vi.mocked(mockClient.post).mockResolvedValue({
        ...mockTask,
        done: true,
        done_at: '2025-01-01T12:00:00Z',
      });

      const result = await taskTools.completeTask(input, userContext);

      expect(result.success).toBe(true);
      expect(result.message).toContain('marked as complete');
      expect(mockClient.post).toHaveBeenCalledWith('/api/v1/tasks/1', {
        done: true,
      });
    });
  });

  describe('deleteTask', () => {
    it('should delete a task successfully', async () => {
      const input = {
        id: 1,
      };

      vi.mocked(mockClient.delete).mockResolvedValue(undefined);

      const result = await taskTools.deleteTask(input, userContext);

      expect(result.success).toBe(true);
      expect(result.message).toContain('deleted successfully');
      expect(mockClient.delete).toHaveBeenCalledWith('/api/v1/tasks/1');
    });
  });

  describe('moveTask', () => {
    it('should move a task to another project', async () => {
      const input = {
        id: 1,
        project_id: 2,
      };

      vi.mocked(mockClient.post).mockResolvedValue({
        ...mockTask,
        project_id: 2,
      });

      const result = await taskTools.moveTask(input, userContext);

      expect(result.success).toBe(true);
      expect(result.message).toContain('moved to project 2');
      expect(mockClient.post).toHaveBeenCalledWith('/api/v1/tasks/1', {
        project_id: 2,
      });
    });

    it('should validate input with Zod schema', () => {
      const invalidInput = {
        id: 1,
        project_id: 0, // Zero should fail (not positive)
      };

      const result = MoveTaskSchema.safeParse(invalidInput);
      expect(result.success).toBe(false);
    });
  });
});
