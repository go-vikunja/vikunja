import { VikunjaClient } from './client.js';
import { VikunjaTask } from './types.js';

/**
 * Input for creating a task
 */
export interface CreateTaskInput {
  title: string;
  description?: string;
  due_date?: string;
  priority?: number;
  labels?: number[];
  assignees?: number[];
}

/**
 * Input for updating a task
 */
export interface UpdateTaskInput {
  title?: string;
  description?: string;
  done?: boolean;
  due_date?: string | null;
  priority?: number;
  labels?: number[];
  assignees?: number[];
}

/**
 * Task API methods
 */
export class TaskAPI {
  constructor(private client: VikunjaClient) {}

  /**
   * Get all tasks
   * @param page - Page number (default: 1)
   * @returns List of tasks
   */
  async getAllTasks(page = 1): Promise<VikunjaTask[]> {
    return this.client.get<VikunjaTask[]>('/api/v1/tasks/all', { page });
  }

  /**
   * Get tasks for a specific project
   * @param projectId - Project ID
   * @param page - Page number (default: 1)
   * @returns List of tasks
   */
  async getProjectTasks(projectId: number, page = 1): Promise<VikunjaTask[]> {
    return this.client.get<VikunjaTask[]>(`/api/v1/projects/${projectId}/tasks`, { page });
  }

  /**
   * Get a single task by ID
   * @param id - Task ID
   * @returns Task details
   */
  async getTask(id: number): Promise<VikunjaTask> {
    return this.client.get<VikunjaTask>(`/api/v1/tasks/${id}`);
  }

  /**
   * Create a new task
   * @param projectId - Project ID
   * @param data - Task data
   * @returns Created task
   */
  async createTask(projectId: number, data: CreateTaskInput): Promise<VikunjaTask> {
    const taskData = { ...data, project_id: projectId };
    return this.client.put<VikunjaTask>(`/api/v1/projects/${projectId}`, taskData);
  }

  /**
   * Update an existing task
   * @param id - Task ID
   * @param data - Updated task data
   * @returns Updated task
   */
  async updateTask(id: number, data: UpdateTaskInput): Promise<VikunjaTask> {
    return this.client.post<VikunjaTask>(`/api/v1/tasks/${id}`, data);
  }

  /**
   * Delete a task
   * @param id - Task ID
   */
  async deleteTask(id: number): Promise<void> {
    await this.client.delete(`/api/v1/tasks/${id}`);
  }

  /**
   * Bulk update tasks
   * @param taskIds - Array of task IDs
   * @param data - Data to update
   * @returns Array of updated tasks
   */
  async bulkUpdateTasks(taskIds: number[], data: Partial<VikunjaTask>): Promise<VikunjaTask[]> {
    const updates = taskIds.map((id) => ({ id, ...data }));
    return this.client.post<VikunjaTask[]>('/api/v1/tasks/bulk', { tasks: updates });
  }
}
