import { describe, it, expect, beforeEach, vi } from 'vitest';
import {
  AssignmentTools,
  AssignTaskSchema,
  AddLabelSchema,
  CreateLabelSchema,
} from '../../../src/tools/assignments.js';
import { VikunjaClient } from '../../../src/vikunja/client.js';
import { RateLimiter } from '../../../src/ratelimit/limiter.js';
import { UserContext } from '../../../src/auth/types.js';
import { VikunjaTask, VikunjaLabel } from '../../../src/vikunja/types.js';

describe('Assignment Tools', () => {
  let assignmentTools: AssignmentTools;
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

  const mockLabel: VikunjaLabel = {
    id: 1,
    title: 'Test Label',
    description: '',
    hex_color: '#ff0000',
    created: '2025-01-01T00:00:00Z',
    updated: '2025-01-01T00:00:00Z',
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

    assignmentTools = new AssignmentTools(mockClient, mockRateLimiter);

    userContext = {
      userId: 1,
      username: 'testuser',
      email: 'test@example.com',
      token: 'test-token',
    };
  });

  describe('assignTask', () => {
    it('should assign user to task', async () => {
      const input = {
        task_id: 1,
        user_id: 2,
      };

      vi.mocked(mockClient.put).mockResolvedValue(mockTask);

      const result = await assignmentTools.assignTask(input, userContext);

      expect(result.success).toBe(true);
      expect(result.message).toContain('assigned to task');
      expect(mockClient.put).toHaveBeenCalledWith('/api/v1/tasks/1/assignees', {
        user_id: 2,
      });
    });

    it('should validate input with Zod schema', () => {
      const invalidInput = {
        task_id: 0,
        user_id: 2,
      };

      const result = AssignTaskSchema.safeParse(invalidInput);
      expect(result.success).toBe(false);
    });
  });

  describe('unassignTask', () => {
    it('should remove user from task', async () => {
      const input = {
        task_id: 1,
        user_id: 2,
      };

      vi.mocked(mockClient.delete).mockResolvedValue(mockTask);

      const result = await assignmentTools.unassignTask(input, userContext);

      expect(result.success).toBe(true);
      expect(result.message).toContain('removed from task');
      expect(mockClient.delete).toHaveBeenCalledWith('/api/v1/tasks/1/assignees/2');
    });
  });

  describe('addLabel', () => {
    it('should add label to task', async () => {
      const input = {
        task_id: 1,
        label_id: 1,
      };

      vi.mocked(mockClient.put).mockResolvedValue(mockTask);

      const result = await assignmentTools.addLabel(input, userContext);

      expect(result.success).toBe(true);
      expect(result.message).toContain('added to task');
      expect(mockClient.put).toHaveBeenCalledWith('/api/v1/tasks/1/labels', {
        label_id: 1,
      });
    });

    it('should validate input with Zod schema', () => {
      const invalidInput = {
        task_id: 1,
        label_id: -1,
      };

      const result = AddLabelSchema.safeParse(invalidInput);
      expect(result.success).toBe(false);
    });
  });

  describe('removeLabel', () => {
    it('should remove label from task', async () => {
      const input = {
        task_id: 1,
        label_id: 1,
      };

      vi.mocked(mockClient.delete).mockResolvedValue(mockTask);

      const result = await assignmentTools.removeLabel(input, userContext);

      expect(result.success).toBe(true);
      expect(result.message).toContain('removed from task');
      expect(mockClient.delete).toHaveBeenCalledWith('/api/v1/tasks/1/labels/1');
    });
  });

  describe('createLabel', () => {
    it('should create a new label', async () => {
      const input = {
        title: 'New Label',
        description: 'Label description',
        hex_color: '#00ff00',
      };

      vi.mocked(mockClient.put).mockResolvedValue(mockLabel);

      const result = await assignmentTools.createLabel(input, userContext);

      expect(result.success).toBe(true);
      expect(result.message).toContain('created successfully');
      expect(result.label).toEqual(mockLabel);
      expect(mockClient.put).toHaveBeenCalledWith('/api/v1/labels', input);
    });

    it('should validate input with Zod schema', () => {
      const invalidInput = {
        title: 'New Label',
        hex_color: 'invalid-color',
      };

      const result = CreateLabelSchema.safeParse(invalidInput);
      expect(result.success).toBe(false);
    });
  });
});
