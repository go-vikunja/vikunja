import { VikunjaClient } from './client.js';
import { VikunjaProject } from './types.js';

/**
 * Input for creating a project
 */
export interface CreateProjectInput {
  title: string;
  description?: string;
  hex_color?: string;
  parent_project_id?: number;
}

/**
 * Input for updating a project
 */
export interface UpdateProjectInput {
  title?: string;
  description?: string;
  hex_color?: string;
  is_archived?: boolean;
  parent_project_id?: number;
}

/**
 * Project API methods
 */
export class ProjectAPI {
  constructor(private client: VikunjaClient) {}

  /**
   * Get all projects
   * @param page - Page number (default: 1)
   * @returns List of projects
   */
  async getProjects(page = 1): Promise<VikunjaProject[]> {
    return this.client.get<VikunjaProject[]>('/api/v1/projects', { page });
  }

  /**
   * Get a single project by ID
   * @param id - Project ID
   * @returns Project details
   */
  async getProject(id: number): Promise<VikunjaProject> {
    return this.client.get<VikunjaProject>(`/api/v1/projects/${id}`);
  }

  /**
   * Create a new project
   * @param data - Project data
   * @returns Created project
   */
  async createProject(data: CreateProjectInput): Promise<VikunjaProject> {
    return this.client.post<VikunjaProject>('/api/v1/projects', data);
  }

  /**
   * Update an existing project
   * @param id - Project ID
   * @param data - Updated project data
   * @returns Updated project
   */
  async updateProject(id: number, data: UpdateProjectInput): Promise<VikunjaProject> {
    return this.client.put<VikunjaProject>(`/api/v1/projects/${id}`, data);
  }

  /**
   * Delete a project
   * @param id - Project ID
   */
  async deleteProject(id: number): Promise<void> {
    await this.client.delete(`/api/v1/projects/${id}`);
  }
}
