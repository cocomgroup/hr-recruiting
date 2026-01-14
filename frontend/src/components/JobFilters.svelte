<script lang="ts">
  import { jobsStore, departments, locations, jobTypes } from '../stores/jobs';

  let searchTerm = '';
  let selectedDepartment = '';
  let selectedType = '';
  let selectedLocation = '';

  $: {
    jobsStore.setFilter('search', searchTerm);
    jobsStore.setFilter('department', selectedDepartment);
    jobsStore.setFilter('type', selectedType);
    jobsStore.setFilter('location', selectedLocation);
  }

  function clearAllFilters() {
    searchTerm = '';
    selectedDepartment = '';
    selectedType = '';
    selectedLocation = '';
    jobsStore.clearFilters();
  }

  $: hasActiveFilters = searchTerm || selectedDepartment || selectedType || selectedLocation;
</script>

<div class="bg-white rounded-lg shadow-sm border border-gray-200 p-6 mb-6">
  <div class="space-y-4">
    <!-- Search -->
    <div>
      <label for="search" class="label">Search</label>
      <div class="relative">
        <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
          <svg class="h-5 w-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
        </div>
        <input
          id="search"
          type="text"
          bind:value={searchTerm}
          placeholder="Search by job title, keyword, or department..."
          class="input pl-10"
        />
      </div>
    </div>

    <!-- Filter Grid -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
      <!-- Department -->
      <div>
        <label for="department" class="label">Department</label>
        <select id="department" bind:value={selectedDepartment} class="input">
          <option value="">All Departments</option>
          {#each $departments as dept}
            <option value={dept}>{dept}</option>
          {/each}
        </select>
      </div>

      <!-- Job Type -->
      <div>
        <label for="type" class="label">Job Type</label>
        <select id="type" bind:value={selectedType} class="input">
          <option value="">All Types</option>
          {#each $jobTypes as type}
            <option value={type}>{type.replace('-', ' ')}</option>
          {/each}
        </select>
      </div>

      <!-- Location -->
      <div>
        <label for="location" class="label">Location</label>
        <select id="location" bind:value={selectedLocation} class="input">
          <option value="">All Locations</option>
          {#each $locations as loc}
            <option value={loc}>{loc}</option>
          {/each}
        </select>
      </div>
    </div>

    <!-- Clear Filters -->
    {#if hasActiveFilters}
      <div class="flex justify-end">
        <button on:click={clearAllFilters} class="btn btn-ghost text-sm">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
          Clear all filters
        </button>
      </div>
    {/if}
  </div>
</div>
