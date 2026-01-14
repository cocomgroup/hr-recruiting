<script lang="ts">
  import { onMount } from 'svelte';
  import { jobsStore, filteredJobs } from '../stores/jobs';
  import JobCard from '../components/JobCard.svelte';
  import JobFilters from '../components/JobFilters.svelte';

  onMount(() => {
    jobsStore.fetchJobs();
  });

  $: loading = $jobsStore.loading;
  $: error = $jobsStore.error;
  $: jobs = $filteredJobs;
</script>

<div class="min-h-screen bg-gray-50">
  <!-- Hero Section -->
  <div class="bg-gradient-to-br from-primary-600 to-primary-800 text-white">
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-16 sm:py-20">
      <div class="text-center">
        <h1 class="text-4xl sm:text-5xl lg:text-6xl font-bold mb-6">
          Join Our Team
        </h1>
        <p class="text-xl sm:text-2xl text-primary-100 mb-8 max-w-3xl mx-auto">
          Build your career with us. Explore opportunities and make an impact.
        </p>
        <div class="flex justify-center items-center space-x-6 text-primary-100">
          <div class="text-center">
            <div class="text-3xl font-bold">{$jobsStore.jobs.length}+</div>
            <div class="text-sm">Open Positions</div>
          </div>
          <div class="w-px h-12 bg-primary-400"></div>
          <div class="text-center">
            <div class="text-3xl font-bold">50+</div>
            <div class="text-sm">Countries</div>
          </div>
          <div class="w-px h-12 bg-primary-400"></div>
          <div class="text-center">
            <div class="text-3xl font-bold">5000+</div>
            <div class="text-sm">Employees</div>
          </div>
        </div>
      </div>
    </div>
  </div>

  <!-- Main Content -->
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
    <!-- Filters -->
    <JobFilters />

    <!-- Results -->
    {#if loading}
      <div class="flex justify-center items-center py-20">
        <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600"></div>
      </div>
    {:else if error}
      <div class="card bg-red-50 border-red-200 text-center py-12">
        <svg class="w-12 h-12 text-red-400 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <h3 class="text-lg font-semibold text-red-900 mb-2">Error Loading Jobs</h3>
        <p class="text-red-700 mb-4">{error}</p>
        <button on:click={() => jobsStore.fetchJobs()} class="btn btn-primary">
          Try Again
        </button>
      </div>
    {:else if jobs.length === 0}
      <div class="card text-center py-12">
        <svg class="w-16 h-16 text-gray-300 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
        </svg>
        <h3 class="text-lg font-semibold text-gray-900 mb-2">No jobs found</h3>
        <p class="text-gray-600 mb-4">Try adjusting your filters to see more results.</p>
        <button on:click={() => jobsStore.clearFilters()} class="btn btn-secondary">
          Clear Filters
        </button>
      </div>
    {:else}
      <div class="mb-4 flex justify-between items-center">
        <p class="text-gray-600">
          Showing <span class="font-semibold">{jobs.length}</span> {jobs.length === 1 ? 'job' : 'jobs'}
        </p>
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {#each jobs as job (job.id)}
          <JobCard {job} />
        {/each}
      </div>
    {/if}
  </div>

  <!-- Why Join Us Section -->
  <div class="bg-white border-t border-gray-200 mt-16">
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-16">
      <h2 class="text-3xl font-bold text-center mb-12">Why Join Us?</h2>
      <div class="grid grid-cols-1 md:grid-cols-3 gap-8">
        <div class="text-center">
          <div class="w-16 h-16 bg-primary-100 rounded-full flex items-center justify-center mx-auto mb-4">
            <svg class="w-8 h-8 text-primary-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
            </svg>
          </div>
          <h3 class="text-xl font-semibold mb-2">Innovation First</h3>
          <p class="text-gray-600">
            Work on cutting-edge projects that shape the future of technology.
          </p>
        </div>

        <div class="text-center">
          <div class="w-16 h-16 bg-primary-100 rounded-full flex items-center justify-center mx-auto mb-4">
            <svg class="w-8 h-8 text-primary-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
            </svg>
          </div>
          <h3 class="text-xl font-semibold mb-2">Collaborative Culture</h3>
          <p class="text-gray-600">
            Join a diverse team of talented professionals from around the world.
          </p>
        </div>

        <div class="text-center">
          <div class="w-16 h-16 bg-primary-100 rounded-full flex items-center justify-center mx-auto mb-4">
            <svg class="w-8 h-8 text-primary-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4M7.835 4.697a3.42 3.42 0 001.946-.806 3.42 3.42 0 014.438 0 3.42 3.42 0 001.946.806 3.42 3.42 0 013.138 3.138 3.42 3.42 0 00.806 1.946 3.42 3.42 0 010 4.438 3.42 3.42 0 00-.806 1.946 3.42 3.42 0 01-3.138 3.138 3.42 3.42 0 00-1.946.806 3.42 3.42 0 01-4.438 0 3.42 3.42 0 00-1.946-.806 3.42 3.42 0 01-3.138-3.138 3.42 3.42 0 00-.806-1.946 3.42 3.42 0 010-4.438 3.42 3.42 0 00.806-1.946 3.42 3.42 0 013.138-3.138z" />
            </svg>
          </div>
          <h3 class="text-xl font-semibold mb-2">Great Benefits</h3>
          <p class="text-gray-600">
            Competitive compensation, health coverage, and flexible work arrangements.
          </p>
        </div>
      </div>
    </div>
  </div>
</div>
