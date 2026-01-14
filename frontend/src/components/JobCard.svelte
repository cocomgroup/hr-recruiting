<script lang="ts">
  import { Link } from 'svelte-routing';
  import type { Job } from '../lib/api';

  export let job: Job;

  function formatSalary(range: Job['salary_range']) {
    if (!range) return null;
    const formatter = new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: range.currency || 'USD',
      minimumFractionDigits: 0,
      maximumFractionDigits: 0,
    });
    return `${formatter.format(range.min)} - ${formatter.format(range.max)}`;
  }

  function formatDate(dateString: string) {
    const date = new Date(dateString);
    const now = new Date();
    const diffTime = Math.abs(now.getTime() - date.getTime());
    const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));

    if (diffDays === 0) return 'Today';
    if (diffDays === 1) return 'Yesterday';
    if (diffDays < 7) return `${diffDays} days ago`;
    if (diffDays < 30) return `${Math.floor(diffDays / 7)} weeks ago`;
    return `${Math.floor(diffDays / 30)} months ago`;
  }

  const typeColors = {
    'full-time': 'badge-primary',
    'part-time': 'badge-success',
    'contract': 'badge-warning',
    'internship': 'badge-gray',
  };
</script>

<Link to={`/jobs/${job.id}`}>
  <div class="card hover:shadow-md transition-shadow duration-200 cursor-pointer group">
    <div class="flex justify-between items-start mb-4">
      <div class="flex-1">
        <h3 class="text-lg font-semibold text-gray-900 group-hover:text-primary-600 transition-colors mb-1">
          {job.title}
        </h3>
        <div class="flex items-center space-x-3 text-sm text-gray-600">
          <span class="flex items-center">
            <svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 13.255A23.931 23.931 0 0112 15c-3.183 0-6.22-.62-9-1.745M16 6V4a2 2 0 00-2-2h-4a2 2 0 00-2 2v2m4 6h.01M5 20h14a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
            </svg>
            {job.department}
          </span>
          <span class="flex items-center">
            <svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
            </svg>
            {job.location}
          </span>
        </div>
      </div>
      <span class={`badge ${typeColors[job.type]}`}>
        {job.type.replace('-', ' ')}
      </span>
    </div>

    <p class="text-gray-600 text-sm mb-4 line-clamp-2">
      {job.description}
    </p>

    <div class="flex items-center justify-between pt-4 border-t border-gray-100">
      <div class="flex items-center space-x-4 text-sm text-gray-500">
        {#if job.salary_range}
          <span class="font-medium text-gray-700">
            {formatSalary(job.salary_range)}
          </span>
        {/if}
        <span>Posted {formatDate(job.posted_at)}</span>
      </div>
      <span class="text-primary-600 font-medium text-sm group-hover:text-primary-700 flex items-center">
        View Details
        <svg class="w-4 h-4 ml-1 group-hover:translate-x-1 transition-transform" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
        </svg>
      </span>
    </div>
  </div>
</Link>

<style>
  .line-clamp-2 {
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }
</style>
