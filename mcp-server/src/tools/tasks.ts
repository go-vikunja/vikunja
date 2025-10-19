import { z } from 'zod';
import { VikunjaClient } from '../vikunja/client.js';
import { RateLimiter } from '../ratelimit/limiter.js';
import { UserContext } from '../auth/types.js';
import { VikunjaTask } from '../vikunja/types.js';
import { logger } from '../utils/logger.js';

/**
 * Input schemas for task tools
 */
export const CreateTaskSchema = z.object({
  project_id: z.number().int().positive(),
  title: z.string().min(1).max(500),
  description: z.string().optional(),
  due_date: z.string().optional(),
  priority: z.number().int().min(0).max(5).optional(),
  labels: z.array(z.number().int().positive()).optional(),
  assignees: z.array(z.number().int().positive()).optional(),
});

export const UpdateTaskSchema = z.object({
  id: z.number().int().positive(),
  title: z.string().min(1).max(500).optional(),
  description: z.string().optional(),
  done: z.boolean().optional(),
  due_date: z.string().nullable().optional(),
  priority: z.number().int().min(0).max(5).optional(),
  labels: z.array(z.number().int().positive()).optional(),
  assignees: z.array(z.number().int().positive()).optional(),
});

export const CompleteTaskSchema = z.object({
  id: z.number().int().positive(),
});

export const DeleteTaskSchema = z.object({
  id: z.number().int().positive(),
});

export const MoveTaskSchema = z.object({
  id: z.number().int().positive(),
  project_id: z.number().int().positive(),
});

export type CreateTaskInput = z.infer<typeof CreateTaskSchema>;
export type UpdateTaskInput = z.infer<typeof UpdateTaskSchema>;
export type CompleteTaskInput = z.infer<typeof CompleteTaskSchema>;
export type DeleteTaskInput = z.infer<typeof DeleteTaskSchema>;
export type MoveTaskInput = z.infer<typeof MoveTaskSchema>;

/**
 * Tool result for task operations
 */
export interface TaskToolResult {
  success: boolean;
  message: string;
  task?: VikunjaTask;
  taskId?: number;
  error?: string;
}

/**
 * Task management tools for MCP protocol
 */
export class TaskTools {
  constructor(
    private client: VikunjaClient,
    private rateLimiter: RateLimiter
  ) {}

  /**
   * Create a new task in a project
   */
  async createTask(
    input: CreateTaskInput,
    userContext: UserContext
  ): Promise<TaskToolResult> {
    try {
      // Rate limiting check
      await this.rateLimiter.checkLimit(userContext.token);

      // Set auth token
      this.client.setToken(userContext.token);

      // Create task
      const task = await this.client.put<VikunjaTask>(
        `/api/v1/projects/${input.project_id}`,
        input
      );

      logger.info('Task created', {
        taskId: task.id,
        projectId: input.project_id,
        userId: userContext.userId,
      });

      return {
        success: true,
        message: `Task "${task.title}" created successfully with ID ${task.id}`,
        task,
        taskId: task.id,
      };
    } catch (error) {
      logger.error('Failed to create task', { error, userId: userContext.userId });
      return {
        success: false,
        message: 'Failed to create task',
        error: error instanceof Error ? error.message : String(error),
      };
    }
  }

  /**
   * Update an existing task
   */
  async updateTask(
    input: UpdateTaskInput,
    userContext: UserContext
  ): Promise<TaskToolResult> {
    try {
      // Rate limiting check
      await this.rateLimiter.checkLimit(userContext.token);

      // Set auth token
      this.client.setToken(userContext.token);

      // Extract ID and update data
      const { id, ...updateData } = input;

      // Update task
      const task = await this.client.post<VikunjaTask>(`/api/v1/tasks/${id}`, updateData);

      logger.info('Task updated', {
        taskId: task.id,
        userId: userContext.userId,
      });

      return {
        success: true,
        message: `Task "${task.title}" updated successfully`,
        task,
        taskId: task.id,
      };
    } catch (error) {
      logger.error('Failed to update task', { error, userId: userContext.userId });
      return {
        success: false,
        message: 'Failed to update task',
        error: error instanceof Error ? error.message : String(error),
      };
    }
  }

  /**
   * Mark a task as complete
   */
  async completeTask(
    input: CompleteTaskInput,
    userContext: UserContext
  ): Promise<TaskToolResult> {
    try {
      // Rate limiting check
      await this.rateLimiter.checkLimit(userContext.token);

      // Set auth token
      this.client.setToken(userContext.token);

      // Complete task
      const task = await this.client.post<VikunjaTask>(`/api/v1/tasks/${input.id}`, {
        done: true,
      });

      logger.info('Task completed', {
        taskId: task.id,
        userId: userContext.userId,
      });

      return {
        success: true,
        message: `Task "${task.title}" marked as complete`,
        task,
        taskId: task.id,
      };
    } catch (error) {
      logger.error('Failed to complete task', { error, userId: userContext.userId });
      return {
        success: false,
        message: 'Failed to complete task',
        error: error instanceof Error ? error.message : String(error),
      };
    }
  }

  /**
   * Delete a task
   */
  async deleteTask(
    input: DeleteTaskInput,
    userContext: UserContext
  ): Promise<TaskToolResult> {
    try {
      // Rate limiting check
      await this.rateLimiter.checkLimit(userContext.token);

      // Set auth token
      this.client.setToken(userContext.token);

      // Delete task
      await this.client.delete(`/api/v1/tasks/${input.id}`);

      logger.info('Task deleted', {
        taskId: input.id,
        userId: userContext.userId,
      });

      return {
        success: true,
        message: `Task with ID ${input.id} deleted successfully`,
        taskId: input.id,
      };
    } catch (error) {
      logger.error('Failed to delete task', { error, userId: userContext.userId });
      return {
        success: false,
        message: 'Failed to delete task',
        error: error instanceof Error ? error.message : String(error),
      };
    }
  }

  /**
   * Move a task to a different project
   */
  async moveTask(
    input: MoveTaskInput,
    userContext: UserContext
  ): Promise<TaskToolResult> {
    try {
      // Rate limiting check
      await this.rateLimiter.checkLimit(userContext.token);

      // Set auth token
      this.client.setToken(userContext.token);

      // Move task (update project_id)
      const task = await this.client.post<VikunjaTask>(`/api/v1/tasks/${input.id}`, {
        project_id: input.project_id,
      });

      logger.info('Task moved', {
        taskId: task.id,
        newProjectId: input.project_id,
        userId: userContext.userId,
      });

      return {
        success: true,
        message: `Task "${task.title}" moved to project ${input.project_id}`,
        task,
        taskId: task.id,
      };
    } catch (error) {
      logger.error('Failed to move task', { error, userId: userContext.userId });
      return {
        success: false,
        message: 'Failed to move task',
        error: error instanceof Error ? error.message : String(error),
      };
    }
  }
}
