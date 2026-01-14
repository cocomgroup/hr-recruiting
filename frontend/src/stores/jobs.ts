import { writable, derived } from 'svelte/store';
import type { Job } from '../lib/api';
import { jobsAPI } from '../lib/api';

interface JobsState {
  jobs: Job[];
  loading: boolean;
  error: string | null;
  filters: {
    department: string;
    type: string;
    location: string;
    search: string;
  };
}

function createJobsStore() {
  const { subscribe, set, update } = writable<JobsState>({
    jobs: [],
    loading: false,
    error: null,
    filters: {
      department: '',
      type: '',
      location: '',
      search: '',
    },
  });

  return {
    subscribe,
    
    async fetchJobs() {
      update(state => ({ ...state, loading: true, error: null }));
      
      try {
        const jobs = await jobsAPI.list();
        update(state => ({ ...state, jobs, loading: false }));
      } catch (error) {
        update(state => ({
          ...state,
          loading: false,
          error: error instanceof Error ? error.message : 'Failed to fetch jobs',
        }));
      }
    },

    setFilter(key: keyof JobsState['filters'], value: string) {
      update(state => ({
        ...state,
        filters: { ...state.filters, [key]: value },
      }));
    },

    clearFilters() {
      update(state => ({
        ...state,
        filters: {
          department: '',
          type: '',
          location: '',
          search: '',
        },
      }));
    },
  };
}

export const jobsStore = createJobsStore();

// Derived store for filtered jobs
export const filteredJobs = derived(jobsStore, $jobsStore => {
  let filtered = $jobsStore.jobs;

  if ($jobsStore.filters.department) {
    filtered = filtered.filter(job => 
      job.department.toLowerCase() === $jobsStore.filters.department.toLowerCase()
    );
  }

  if ($jobsStore.filters.type) {
    filtered = filtered.filter(job => 
      job.type.toLowerCase() === $jobsStore.filters.type.toLowerCase()
    );
  }

  if ($jobsStore.filters.location) {
    filtered = filtered.filter(job => 
      job.location.toLowerCase().includes($jobsStore.filters.location.toLowerCase())
    );
  }

  if ($jobsStore.filters.search) {
    const searchLower = $jobsStore.filters.search.toLowerCase();
    filtered = filtered.filter(job =>
      job.title.toLowerCase().includes(searchLower) ||
      job.description.toLowerCase().includes(searchLower) ||
      job.department.toLowerCase().includes(searchLower)
    );
  }

  return filtered;
});

// Derived stores for unique filter values
export const departments = derived(jobsStore, $jobsStore => {
  const depts = new Set($jobsStore.jobs.map(job => job.department));
  return Array.from(depts).sort();
});

export const locations = derived(jobsStore, $jobsStore => {
  const locs = new Set($jobsStore.jobs.map(job => job.location));
  return Array.from(locs).sort();
});

export const jobTypes = derived(jobsStore, $jobsStore => {
  const types = new Set($jobsStore.jobs.map(job => job.type));
  return Array.from(types).sort();
});
