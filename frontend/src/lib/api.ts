// API client for HR-Recruiting backend

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8081/api/v1';

export interface Job {
  id: string;
  title: string;
  department: string;
  location: string;
  employmentType: string;
  experienceLevel?: string;
  salaryRange?: {
    min: number;
    max: number;
    currency: string;
  };
  description: string;
  requirements?: string;
  responsibilities?: string;
  benefits?: string;
  skills?: string[];
  status: string;
  postedDate?: string;
  closingDate?: string;
  applicationCount: number;
  viewCount: number;
  remoteWork?: boolean;
  urgentHiring?: boolean;
  createdAt?: string;
  updatedAt?: string;
}

export interface Application {
  id: string;
  jobId: string;
  firstName: string;
  lastName: string;
  email: string;
  phone: string;
  resumeUrl: string;
  coverLetter?: string;
  linkedinUrl?: string;
  portfolioUrl?: string;
  status: string;
  appliedDate: string;
}

export interface ApplicationSubmission {
  jobId: string;
  firstName: string;
  lastName: string;
  email: string;
  phone: string;
  resumeUrl: string;
  currentLocation: string;  // Required by backend
  availability: string;     // Required by backend
  coverLetter?: string;
  linkedinUrl?: string;
  portfolioUrl?: string;
  yearsOfExperience?: number;
  expectedSalary?: number;
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

  const data = await response.json();
  
  // The GraphQL response is wrapped in { data: { jobs: [...] } }
  // Extract just the inner data
  if (data.data) {
    return data.data;
  }
  
  return data;
}

// Job API
export const jobsAPI = {
  async list(filters?: { department?: string; employmentType?: string; location?: string }): Promise<Job[]> {
    const params = new URLSearchParams();
    if (filters?.department) params.append('department', filters.department);
    if (filters?.employmentType) params.append('employmentType', filters.employmentType);
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
  async getPresignedURL(filename: string, contentType: string): Promise<{ uploadUrl: string; url: string; key: string }> {
    const data = await fetchAPI('/upload/presigned-url', {
      method: 'POST',
      body: JSON.stringify({ filename, contentType }),
    });
    return data;
  },

  async uploadFile(file: File): Promise<string> {
    // Get presigned URL
    const { uploadUrl, url } = await this.getPresignedURL(file.name, file.type);

    // Upload to S3 using the presigned URL
    const uploadResponse = await fetch(uploadUrl, {
      method: 'PUT',
      body: file,
      headers: {
        'Content-Type': file.type,
      },
    });

    if (!uploadResponse.ok) {
      throw new Error('Failed to upload file');
    }

    // Return the final S3 URL (not the presigned one)
    return url;
  },
};

export { APIError };
