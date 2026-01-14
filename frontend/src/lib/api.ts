// API client for HR-Recruiting backend

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1';

export interface Job {
  id: string;
  title: string;
  department: string;
  location: string;
  type: 'full-time' | 'part-time' | 'contract' | 'internship';
  salary_range?: {
    min: number;
    max: number;
    currency: string;
  };
  description: string;
  requirements: string[];
  responsibilities: string[];
  benefits: string[];
  posted_at: string;
  status: 'draft' | 'open' | 'closed';
  views: number;
  applications_count: number;
}

export interface Application {
  id: string;
  job_id: string;
  first_name: string;
  last_name: string;
  email: string;
  phone: string;
  resume_url: string;
  cover_letter?: string;
  linkedin_url?: string;
  portfolio_url?: string;
  status: 'submitted' | 'reviewing' | 'interviewing' | 'offered' | 'rejected';
  submitted_at: string;
}

export interface ApplicationSubmission {
  job_id: string;
  first_name: string;
  last_name: string;
  email: string;
  phone: string;
  resume_url: string;
  cover_letter?: string;
  linkedin_url?: string;
  portfolio_url?: string;
}

class APIError extends Error {
  constructor(public status: number, message: string) {
    super(message);
    this.name = 'APIError';
  }
}

async function fetchAPI(endpoint: string, options: RequestInit = {}) {
  const url = `${API_BASE_URL}${endpoint}`;
  
  const response = await fetch(url, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options.headers,
    },
  });

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new APIError(
      response.status,
      errorData.error?.message || `Request failed with status ${response.status}`
    );
  }

  return response.json();
}

// Job API
export const jobsAPI = {
  async list(filters?: { department?: string; type?: string; location?: string }): Promise<Job[]> {
    const params = new URLSearchParams();
    if (filters?.department) params.append('department', filters.department);
    if (filters?.type) params.append('type', filters.type);
    if (filters?.location) params.append('location', filters.location);
    
    const queryString = params.toString();
    const endpoint = queryString ? `/jobs?${queryString}` : '/jobs';
    
    const data = await fetchAPI(endpoint);
    return data.jobs || [];
  },

  async get(id: string): Promise<Job> {
    const data = await fetchAPI(`/jobs/${id}`);
    return data.job;
  },

  async incrementView(id: string): Promise<void> {
    await fetchAPI(`/jobs/${id}/view`, {
      method: 'POST',
    });
  },
};

// Application API
export const applicationsAPI = {
  async submit(application: ApplicationSubmission): Promise<Application> {
    const data = await fetchAPI('/applications', {
      method: 'POST',
      body: JSON.stringify(application),
    });
    return data.application;
  },
};

// Upload API
export const uploadAPI = {
  async getPresignedURL(filename: string, contentType: string): Promise<{ url: string; key: string }> {
    const data = await fetchAPI('/upload/presigned-url', {
      method: 'POST',
      body: JSON.stringify({ filename, content_type: contentType }),
    });
    return data;
  },

  async uploadFile(file: File): Promise<string> {
    // Get presigned URL
    const { url, key } = await this.getPresignedURL(file.name, file.type);

    // Upload to S3
    const uploadResponse = await fetch(url, {
      method: 'PUT',
      body: file,
      headers: {
        'Content-Type': file.type,
      },
    });

    if (!uploadResponse.ok) {
      throw new Error('Failed to upload file');
    }

    // Return the S3 URL (without query params)
    return url.split('?')[0];
  },
};

export { APIError };
