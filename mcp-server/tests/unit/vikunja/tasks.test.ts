import { describe, it, expect, beforeEach, vi } from 'vitest';
import { VikunjaClient } from '../../../src/vikunja/client.js';
import { TaskAPI } from '../../../src/vikunja/tasks.js';
import { VikunjaTask } from '../../../src/vikunja/types.js';

describe('Task API Methods', () => {
  let mockClient: VikunjaClient;
  let taskAPI: TaskAPI;

  const mockUser = {
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

    taskAPI = new TaskAPI(mockClient);
  });

  describe('getAllTasks', () => {
    it('should get all tasks', async () => {
      const mockTasks: VikunjaTask[] = [
        {
          id: 1,
          title: 'Task 1',
          description: 'Test task 1',
          done: false,
          done_at: null,
          due_date: '2024-12-31T00:00:00Z',
          priority: 1,
          labels: [],
          assignees: [mockUser],
          project_id: 1,
          created: '2024-01-01T00:00:00Z',
          updated: '2024-01-01T00:00:00Z',
          created_by: mockUser,
        },
        {
          id: 2,
          title: 'Task 2',
          description: 'Test task 2',
          done: true,
          done_at: '2024-01-15T00:00:00Z',
          due_date: null,
          priority: 2,
          labels: [],
          assignees: [],
          project_id: 1,
          created: '2024-01-02T00:00:00Z',
          updated: '2024-01-15T00:00:00Z',
          created_by: mockUser,
        },
      ];

      vi.mocked(mockClient.get).mockResolvedValue(mockTasks);

      const tasks = await taskAPI.getAllTasks();
      expect(tasks).toEqual(mockTasks);
      expect(mockClient.get).toHaveBeenCalledWith('/api/v1/tasks/all', { page: 1 });
    });

    it('should handle pagination', async () => {
      const mockTasks: VikunjaTask[] = [
        {
          id: 51,
          title: 'Task 51',
          description: 'Test task on page 2',
          done: false,
          done_at: null,
          due_date: null,
          priority: 0,
          labels: [],
          assignees: [],
          project_id: 1,
          created: '2024-01-01T00:00:00Z',
          updated: '2024-01-01T00:00:00Z',
          created_by: mockUser,
        },
      ];

      vi.mocked(mockClient.get).mockResolvedValue(mockTasks);

      const tasks = await taskAPI.getAllTasks(2);
      expect(tasks).toEqual(mockTasks);
      expect(mockClient.get).toHaveBeenCalledWith('/api/v1/tasks/all', { page: 2 });
    });
  });

  describe('getProjectTasks', () => {
    it('should get project tasks', async () => {
      const mockTasks: VikunjaTask[] = [
        {
          id: 1,
          title: 'Task 1',
          description: 'Test task 1',
          done: false,
          done_at: null,
          due_date: null,
          priority: 0,
          labels: [],
          assignees: [],
          project_id: 5,
          created: '2024-01-01T00:00:00Z',
          updated: '2024-01-01T00:00:00Z',
          created_by: mockUser,
        },
      ];

      vi.mocked(mockClient.get).mockResolvedValue(mockTasks);

      const tasks = await taskAPI.getProjectTasks(5);
      expect(tasks).toEqual(mockTasks);
      expect(mockClient.get).toHaveBeenCalledWith('/api/v1/projects/5/tasks', { page: 1 });
    });
  });

  describe('getTask', () => {
    it('should get single task', async () => {
      const mockTask: VikunjaTask = {
        id: 1,
        title: 'Task 1',
        description: 'Test task 1',
        done: false,
        done_at: null,
        due_date: '2024-12-31T00:00:00Z',
        priority: 1,
        labels: [],
        assignees: [mockUser],
        project_id: 1,
        created: '2024-01-01T00:00:00Z',
        updated: '2024-01-01T00:00:00Z',
        created_by: mockUser,
      };

      vi.mocked(mockClient.get).mockResolvedValue(mockTask);

      const task = await taskAPI.getTask(1);
      expect(task).toEqual(mockTask);
      expect(mockClient.get).toHaveBeenCalledWith('/api/v1/tasks/1');
    });
  });

  describe('createTask', () => {
    it('should create task', async () => {
      const input = {
        title: 'New Task',
        description: 'A new test task',
        priority: 2,
      };

      const mockTask: VikunjaTask = {
        id: 3,
        title: input.title,
        description: input.description,
        done: false,
        done_at: null,
        due_date: null,
        priority: input.priority,
        labels: [],
        assignees: [],
        project_id: 1,
        created: '2024-01-03T00:00:00Z',
        updated: '2024-01-03T00:00:00Z',
        created_by: mockUser,
      };

      vi.mocked(mockClient.put).mockResolvedValue(mockTask);

      const task = await taskAPI.createTask(1, input);
      expect(task).toEqual(mockTask);
      expect(mockClient.put).toHaveBeenCalledWith('/api/v1/projects/1', {
        ...input,
        project_id: 1,
      });
    });
  });

  describe('updateTask', () => {
    it('should update task', async () => {
      const input = {
        title: 'Updated Task',
        description: 'Updated description',
        done: true,
      };

      const mockTask: VikunjaTask = {
        id: 1,
        title: input.title,
        description: input.description,
        done: input.done,
        done_at: '2024-01-05T00:00:00Z',
        due_date: null,
        priority: 0,
        labels: [],
        assignees: [],
        project_id: 1,
        created: '2024-01-01T00:00:00Z',
        updated: '2024-01-05T00:00:00Z',
        created_by: mockUser,
      };

      vi.mocked(mockClient.post).mockResolvedValue(mockTask);

      const task = await taskAPI.updateTask(1, input);
      expect(task).toEqual(mockTask);
      expect(mockClient.post).toHaveBeenCalledWith('/api/v1/tasks/1', input);
    });
  });

  describe('deleteTask', () => {
    it('should delete task', async () => {
      vi.mocked(mockClient.delete).mockResolvedValue({ message: 'Successfully deleted.' });

      await taskAPI.deleteTask(1);
      expect(mockClient.delete).toHaveBeenCalledWith('/api/v1/tasks/1');
    });
  });

  describe('bulkUpdateTasks', () => {
    it('should bulk update tasks', async () => {
      const taskIds = [1, 2, 3];
      const data = { done: true, priority: 2 };

      const mockTasks: VikunjaTask[] = [
        {
          id: 1,
          title: 'Task 1',
          description: '',
          done: true,
          done_at: '2024-01-05T00:00:00Z',
          due_date: null,
          priority: 2,
          labels: [],
          assignees: [],
          project_id: 1,
          created: '2024-01-01T00:00:00Z',
          updated: '2024-01-05T00:00:00Z',
          created_by: mockUser,
        },
        {
          id: 2,
          title: 'Task 2',
          description: '',
          done: true,
          done_at: '2024-01-05T00:00:00Z',
          due_date: null,
          priority: 2,
          labels: [],
          assignees: [],
          project_id: 1,
          created: '2024-01-02T00:00:00Z',
          updated: '2024-01-05T00:00:00Z',
          created_by: mockUser,
        },
        {
          id: 3,
          title: 'Task 3',
          description: '',
          done: true,
          done_at: '2024-01-05T00:00:00Z',
          due_date: null,
          priority: 2,
          labels: [],
          assignees: [],
          project_id: 1,
          created: '2024-01-03T00:00:00Z',
          updated: '2024-01-05T00:00:00Z',
          created_by: mockUser,
        },
      ];

      vi.mocked(mockClient.post).mockResolvedValue(mockTasks);

      const tasks = await taskAPI.bulkUpdateTasks(taskIds, data);
      expect(tasks).toEqual(mockTasks);
      expect(mockClient.post).toHaveBeenCalledWith('/api/v1/tasks/bulk', {
        tasks: [
          { id: 1, done: true, priority: 2 },
          { id: 2, done: true, priority: 2 },
          { id: 3, done: true, priority: 2 },
        ],
      });
    });
  });
});
