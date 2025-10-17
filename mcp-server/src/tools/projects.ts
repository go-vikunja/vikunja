import { z } from 'zod';
import { VikunjaClient } from '../vikunja/client.js';
import { RateLimiter } from '../ratelimit/limiter.js';
import { UserContext } from '../auth/types.js';
import { VikunjaProject } from '../vikunja/types.js';
import { logger } from '../utils/logger.js';

/**
 * Input schemas for project tools
 */
export const CreateProjectSchema = z.object({
  title: z.string().min(1).max(250),
  description: z.string().optional(),
  hex_color: z.string().regex(/^#[0-9a-fA-F]{6}$/).optional(),
  parent_project_id: z.number().int().positive().optional(),
});

export const UpdateProjectSchema = z.object({
  id: z.number().int().positive(),
  title: z.string().min(1).max(250).optional(),
  description: z.string().optional(),
  hex_color: z.string().regex(/^#[0-9a-fA-F]{6}$/).optional(),
  is_archived: z.boolean().optional(),
  parent_project_id: z.number().int().positive().optional(),
});

export const DeleteProjectSchema = z.object({
  id: z.number().int().positive(),
});

export const ArchiveProjectSchema = z.object({
  id: z.number().int().positive(),
  archived: z.boolean(),
});

export type CreateProjectInput = z.infer<typeof CreateProjectSchema>;
export type UpdateProjectInput = z.infer<typeof UpdateProjectSchema>;
export type DeleteProjectInput = z.infer<typeof DeleteProjectSchema>;
export type ArchiveProjectInput = z.infer<typeof ArchiveProjectSchema>;

/**
 * Tool result for project operations
 */
export interface ProjectToolResult {
  success: boolean;
  message: string;
  project?: VikunjaProject;
  error?: string;
}

/**
 * Project management tools for MCP protocol
 */
export class ProjectTools {
  constructor(
    private client: VikunjaClient,
    private rateLimiter: RateLimiter
  ) {}

  /**
   * Create a new project
   */
  async createProject(
    input: CreateProjectInput,
    userContext: UserContext
  ): Promise<ProjectToolResult> {
    try {
      // Rate limiting check
      await this.rateLimiter.checkLimit(userContext.token);

      // Set auth token
      this.client.setToken(userContext.token);

      // Create project
      const project = await this.client.post<VikunjaProject>('/api/v1/projects', input);

      logger.info('Project created', {
        projectId: project.id,
        userId: userContext.userId,
      });

      return {
        success: true,
        message: `Project "${project.title}" created successfully with ID ${project.id}`,
        project,
      };
    } catch (error) {
      logger.error('Failed to create project', { error, userId: userContext.userId });
      return {
        success: false,
        message: 'Failed to create project',
        error: error instanceof Error ? error.message : String(error),
      };
    }
  }

  /**
   * Update an existing project
   */
  async updateProject(
    input: UpdateProjectInput,
    userContext: UserContext
  ): Promise<ProjectToolResult> {
    try {
      // Rate limiting check
      await this.rateLimiter.checkLimit(userContext.token);

      // Set auth token
      this.client.setToken(userContext.token);

      // Extract ID and update data
      const { id, ...updateData } = input;

      // Update project
      const project = await this.client.put<VikunjaProject>(
        `/api/v1/projects/${id}`,
        updateData
      );

      logger.info('Project updated', {
        projectId: project.id,
        userId: userContext.userId,
      });

      return {
        success: true,
        message: `Project "${project.title}" updated successfully`,
        project,
      };
    } catch (error) {
      logger.error('Failed to update project', { error, userId: userContext.userId });
      return {
        success: false,
        message: 'Failed to update project',
        error: error instanceof Error ? error.message : String(error),
      };
    }
  }

  /**
   * Delete a project
   */
  async deleteProject(
    input: DeleteProjectInput,
    userContext: UserContext
  ): Promise<ProjectToolResult> {
    try {
      // Rate limiting check
      await this.rateLimiter.checkLimit(userContext.token);

      // Set auth token
      this.client.setToken(userContext.token);

      // Delete project
      await this.client.delete(`/api/v1/projects/${input.id}`);

      logger.info('Project deleted', {
        projectId: input.id,
        userId: userContext.userId,
      });

      return {
        success: true,
        message: `Project with ID ${input.id} deleted successfully`,
      };
    } catch (error) {
      logger.error('Failed to delete project', { error, userId: userContext.userId });
      return {
        success: false,
        message: 'Failed to delete project',
        error: error instanceof Error ? error.message : String(error),
      };
    }
  }

  /**
   * Archive or unarchive a project
   */
  async archiveProject(
    input: ArchiveProjectInput,
    userContext: UserContext
  ): Promise<ProjectToolResult> {
    try {
      // Rate limiting check
      await this.rateLimiter.checkLimit(userContext.token);

      // Set auth token
      this.client.setToken(userContext.token);

      // Archive/unarchive project
      const project = await this.client.put<VikunjaProject>(`/api/v1/projects/${input.id}`, {
        is_archived: input.archived,
      });

      logger.info('Project archive status changed', {
        projectId: project.id,
        archived: input.archived,
        userId: userContext.userId,
      });

      return {
        success: true,
        message: `Project "${project.title}" ${input.archived ? 'archived' : 'unarchived'} successfully`,
        project,
      };
    } catch (error) {
      logger.error('Failed to change project archive status', {
        error,
        userId: userContext.userId,
      });
      return {
        success: false,
        message: 'Failed to change project archive status',
        error: error instanceof Error ? error.message : String(error),
      };
    }
  }
}
