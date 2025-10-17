import { describe, it, expect, beforeEach, vi } from 'vitest';
import { TaskResources } from '../../../src/resources/tasks.js';
import { VikunjaClient } from '../../../src/vikunja/client.js';
import { UserContext } from '../../../src/auth/types.js';
import { VikunjaTask } from '../../../src/vikunja/types.js';

describe('Task Resources', () => {
  let taskResources: TaskResources;
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

    taskResources = new TaskResources(mockClient);
  });

  describe('list', () => {
    it('should list all tasks', async () => {
      const mockTasks: VikunjaTask[] = [
        {
          id: 1,
          title: 'Task 1',
          description: 'Description 1',
          done: false,
          done_at: null,
          due_date: null,
          priority: 0,
          labels: [],
          assignees: [],
          project_id: 1,
          created: '2024-01-01',
          updated: '2024-01-01',
          created_by: { id: 1, username: 'testuser', email: 'test@example.com', name: 'Test User', created: '2024-01-01', updated: '2024-01-01' },
        },
      ];

      vi.spyOn(mockClient, 'get').mockResolvedValue(mockTasks);

      const result = await taskResources.list(mockUserContext);

      expect(mockClient.setToken).toHaveBeenCalledWith('test-token');
      expect(mockClient.get).toHaveBeenCalledWith('/api/v1/tasks/all', { page: 1 });
      expect(result).toHaveLength(1);
      expect(result[0].uri).toBe('vikunja://tasks/1');
      expect(result[0].name).toBe('Task 1');
    });

    it('should handle pagination', async () => {
      vi.spyOn(mockClient, 'get').mockResolvedValue([]);

      await taskResources.list(mockUserContext, 3);

      expect(mockClient.get).toHaveBeenCalledWith('/api/v1/tasks/all', { page: 3 });
    });
  });

  describe('read', () => {
    it('should read single task', async () => {
      const mockTask: VikunjaTask = {
        id: 1,
        title: 'Test Task',
        description: 'Test Description',
        done: false,
        done_at: null,
        due_date: '2024-12-31',
        priority: 2,
        labels: [],
        assignees: [],
        project_id: 1,
        created: '2024-01-01',
        updated: '2024-01-01',
        created_by: { id: 1, username: 'testuser', email: 'test@example.com', name: 'Test User', created: '2024-01-01', updated: '2024-01-01' },
      };

      vi.spyOn(mockClient, 'get').mockResolvedValue(mockTask);

      const result = await taskResources.read('vikunja://tasks/1', mockUserContext);

      expect(mockClient.setToken).toHaveBeenCalledWith('test-token');
      expect(mockClient.get).toHaveBeenCalledWith('/api/v1/tasks/1');
      expect(result.contents).toHaveLength(1);
      expect(result.contents[0].uri).toBe('vikunja://tasks/1');
      expect(JSON.parse(result.contents[0].text)).toEqual(mockTask);
    });

    it('should include expanded data', async () => {
      const mockTask: VikunjaTask = {
        id: 1,
        title: 'Test Task',
        description: 'Test Description',
        done: false,
        done_at: null,
        due_date: '2024-12-31',
        priority: 2,
        labels: [{ id: 1, title: 'urgent', description: '', hex_color: '#ff0000', created: '2024-01-01', updated: '2024-01-01' }],
        assignees: [{ id: 2, username: 'assignee', email: 'assignee@example.com', name: 'Assignee', created: '2024-01-01', updated: '2024-01-01' }],
        project_id: 1,
        created: '2024-01-01',
        updated: '2024-01-01',
        created_by: { id: 1, username: 'testuser', email: 'test@example.com', name: 'Test User', created: '2024-01-01', updated: '2024-01-01' },
      };

      vi.spyOn(mockClient, 'get').mockResolvedValue(mockTask);

      const result = await taskResources.read('vikunja://tasks/1', mockUserContext);

      const parsed = JSON.parse(result.contents[0].text);
      expect(parsed.labels).toHaveLength(1);
      expect(parsed.assignees).toHaveLength(1);
      expect(parsed.priority).toBe(2);
    });

    it('should handle not found', async () => {
      vi.spyOn(mockClient, 'get').mockRejectedValue(new Error('Task not found'));

      await expect(
        taskResources.read('vikunja://tasks/999', mockUserContext)
      ).rejects.toThrow('Task not found');
    });

    it('should parse task ID from URI', async () => {
      const mockTask: VikunjaTask = {
        id: 42,
        title: 'Test',
        description: '',
        done: false,
        done_at: null,
        due_date: null,
        priority: 0,
        labels: [],
        assignees: [],
        project_id: 1,
        created: '2024-01-01',
        updated: '2024-01-01',
        created_by: { id: 1, username: 'testuser', email: 'test@example.com', name: 'Test User', created: '2024-01-01', updated: '2024-01-01' },
      };

      vi.spyOn(mockClient, 'get').mockResolvedValue(mockTask);

      await taskResources.read('vikunja://tasks/42', mockUserContext);

      expect(mockClient.get).toHaveBeenCalledWith('/api/v1/tasks/42');
    });
  });
});
