import { writable, derived, type Writable } from 'svelte/store';
import type { Job, Application, JobFilters, Candidate } from '$lib/types';

// Jobs store
function createJobsStore() {
  const { subscribe, set, update }: Writable<Job[]> = writable([]);

  return {
    subscribe,
    set,
    update,
    addJob: (job: Job) => update((jobs) => [...jobs, job]),
    updateJob: (id: string, updates: Partial<Job>) =>
      update((jobs) => jobs.map((job) => (job.id === id ? { ...job, ...updates } : job))),
    removeJob: (id: string) => update((jobs) => jobs.filter((job) => job.id !== id)),
    reset: () => set([]),
  };
}

// Applications store
function createApplicationsStore() {
  const { subscribe, set, update }: Writable<Application[]> = writable([]);

  return {
    subscribe,
    set,
    update,
    addApplication: (application: Application) =>
      update((apps) => [...apps, application]),
    updateApplication: (id: string, updates: Partial<Application>) =>
      update((apps) =>
        apps.map((app) => (app.id === id ? { ...app, ...updates } : app))
      ),
    reset: () => set([]),
  };
}

// Search filters store
function createFiltersStore() {
  const { subscribe, set, update }: Writable<JobFilters> = writable({});

  return {
    subscribe,
    set,
    update,
    updateFilter: <K extends keyof JobFilters>(key: K, value: JobFilters[K]) =>
      update((filters) => ({ ...filters, [key]: value })),
    reset: () => set({}),
  };
}

// Auth store
interface AuthState {
  isAuthenticated: boolean;
  user: any | null;
  token: string | null;
}

function createAuthStore() {
  const { subscribe, set }: Writable<AuthState> = writable({
    isAuthenticated: false,
    user: null,
    token: null,
  });

  return {
    subscribe,
    login: (user: any, token: string) => {
      localStorage.setItem('auth_token', token);
      set({ isAuthenticated: true, user, token });
    },
    logout: () => {
      localStorage.removeItem('auth_token');
      set({ isAuthenticated: false, user: null, token: null });
    },
    checkAuth: () => {
      const token = localStorage.getItem('auth_token');
      if (token) {
        set({ isAuthenticated: true, user: null, token });
      }
    },
  };
}

// Notification store
interface Notification {
  id: string;
  type: 'success' | 'error' | 'info' | 'warning';
  message: string;
  duration?: number;
}

function createNotificationStore() {
  const { subscribe, update }: Writable<Notification[]> = writable([]);

  return {
    subscribe,
    add: (
      type: Notification['type'],
      message: string,
      duration: number = 5000
    ) => {
      const id = Math.random().toString(36).substr(2, 9);
      const notification: Notification = { id, type, message, duration };

      update((notifications) => [...notifications, notification]);

      if (duration > 0) {
        setTimeout(() => {
          update((notifications) =>
            notifications.filter((n) => n.id !== id)
          );
        }, duration);
      }
    },
    remove: (id: string) => {
      update((notifications) => notifications.filter((n) => n.id !== id));
    },
  };
}

// Loading state
export const loading: Writable<boolean> = writable(false);

// Export stores
export const jobs = createJobsStore();
export const applications = createApplicationsStore();
export const searchFilters = createFiltersStore();
export const auth = createAuthStore();
export const notifications = createNotificationStore();

// Derived stores
export const activeJobs = derived(jobs, ($jobs) =>
  $jobs.filter((job) => job.status === 'PUBLISHED')
);

export const filteredJobs = derived(
  [jobs, searchFilters],
  ([$jobs, $filters]) => {
    let filtered = $jobs.filter((job) => job.status === 'PUBLISHED');

    if ($filters.query) {
      const query = $filters.query.toLowerCase();
      filtered = filtered.filter(
        (job) =>
          job.title.toLowerCase().includes(query) ||
          job.description.toLowerCase().includes(query) ||
          job.department.toLowerCase().includes(query)
      );
    }

    if ($filters.departments && $filters.departments.length > 0) {
      filtered = filtered.filter((job) =>
        $filters.departments!.includes(job.department)
      );
    }

    if ($filters.locations && $filters.locations.length > 0) {
      filtered = filtered.filter((job) =>
        $filters.locations!.includes(job.location)
      );
    }

    if ($filters.employmentTypes && $filters.employmentTypes.length > 0) {
      filtered = filtered.filter((job) =>
        $filters.employmentTypes!.includes(job.employmentType)
      );
    }

    if ($filters.remoteWork !== undefined) {
      filtered = filtered.filter((job) => job.remoteWork === $filters.remoteWork);
    }

    if ($filters.salaryMin) {
      filtered = filtered.filter(
        (job) => job.salaryRange && job.salaryRange.max >= $filters.salaryMin!
      );
    }

    return filtered;
  }
);