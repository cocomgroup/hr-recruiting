<script lang="ts">
  import { onMount } from 'svelte';
  
  import { jobsAPI, applicationsAPI, uploadAPI, type Job } from '../lib/api';

  export let id: string;

  let job: Job | null = null;
  let loading = true;
  let error: string | null = null;
  let showApplicationForm = false;

  // Application form state
  let submitting = false;
  let submitError: string | null = null;
  let submitSuccess = false;

  let formData = {
    firstName: '',
    lastName: '',
    email: '',
    phone: '',
    resumeFile: null as File | null,
    coverLetter: '',
    linkedinUrl: '',
    portfolioUrl: '',
    currentLocation: '',
    availability: 'Immediate',
  };

  let resumeUploading = false;

  onMount(async () => {
    try {
      job = await jobsAPI.get(id);
      await jobsAPI.incrementView(id);
      loading = false;
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load job';
      loading = false;
    }
  });

  function handleFileSelect(event: Event) {
    const target = event.target as HTMLInputElement;
    if (target.files && target.files[0]) {
      formData.resumeFile = target.files[0];
    }
  }

  async function handleSubmit() {
    if (!formData.resumeFile) {
      submitError = 'Please upload your resume';
      return;
    }

    submitting = true;
    submitError = null;

    try {
      // Upload resume
      resumeUploading = true;

      // Create form data
      const uploadFormData = new FormData();
      uploadFormData.append('file', formData.resumeFile);

      // Upload through backend
      const uploadResponse = await fetch('http://localhost:8081/api/v1/upload/resume', {
        method: 'POST',
        body: uploadFormData,
      });

      if (!uploadResponse.ok) {
        throw new Error('Failed to upload resume');
      }

      const uploadResult = await uploadResponse.json();
      const resumeUrl = uploadResult.url;
      resumeUploading = false;

      // Submit application
      await applicationsAPI.submit({
        jobId: id,
        firstName: formData.firstName,
        lastName: formData.lastName,
        email: formData.email,
        phone: formData.phone,
        resumeUrl: resumeUrl,
        currentLocation: formData.currentLocation,
        availability: formData.availability,
        coverLetter: formData.coverLetter || undefined,
        linkedinUrl: formData.linkedinUrl || undefined,
        portfolioUrl: formData.portfolioUrl || undefined,
      });

      submitSuccess = true;
    } catch (err) {
      submitError = err instanceof Error ? err.message : 'Failed to submit application';
      submitting = false;
    }
  }

  function formatSalary(range: Job['salaryRange']) {
    if (!range) return null;
    const formatter = new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: range.currency || 'USD',
      minimumFractionDigits: 0,
      maximumFractionDigits: 0,
    });
    return `${formatter.format(range.min)} - ${formatter.format(range.max)}`;
  }

  function formatEmploymentType(type: string): string {
    if (!type) return '';
    // Convert FULL_TIME to Full Time, part-time to Part Time, etc.
    return type.replace(/_/g, ' ').replace(/-/g, ' ').replace(/\b\w/g, l => l.toUpperCase());
  }

  function getTypeBadgeClass(type: string): string {
    if (!type) return 'badge-gray';
    const normalized = type.toLowerCase().replace(/_/g, '-');
    const colorMap: Record<string, string> = {
      'full-time': 'badge-primary',
      'part-time': 'badge-success',
      'contract': 'badge-warning',
      'internship': 'badge-gray',
    };
    return colorMap[normalized] || 'badge-gray';
  }

  function navigate(path: string) {
    if (typeof window !== 'undefined' && (window as any).navigate) {
      (window as any).navigate(path);
    } else {
      window.location.href = path;
    }
  }
</script>

<div class="min-h-screen bg-gray-50">
  {#if loading}
    <div class="flex justify-center items-center min-h-screen">
      <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600"></div>
    </div>
  {:else if error}
    <div class="max-w-3xl mx-auto px-4 py-16 text-center">
      <div class="card bg-red-50 border-red-200">
        <h2 class="text-2xl font-bold text-red-900 mb-4">Job Not Found</h2>
        <p class="text-red-700 mb-6">{error}</p>
        <button on:click={() => navigate('/')} class="btn btn-primary">
          Back to Jobs
        </button>
      </div>
    </div>
  {:else if job}
    <!-- Header -->
    <div class="bg-white border-b border-gray-200">
      <div class="max-w-5xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <button on:click={() => navigate('/')} class="text-gray-600 hover:text-gray-900 mb-4 flex items-center">
          <svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
          </svg>
          Back to Jobs
        </button>

        <div class="flex justify-between items-start">
          <div class="flex-1">
            <h1 class="text-3xl font-bold text-gray-900 mb-2">{job.title}</h1>
            <div class="flex flex-wrap items-center gap-4 text-gray-600 mb-4">
              <span class="flex items-center">
                <svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 13.255A23.931 23.931 0 0112 15c-3.183 0-6.22-.62-9-1.745M16 6V4a2 2 0 00-2-2h-4a2 2 0 00-2 2v2m4 6h.01M5 20h14a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
                </svg>
                {job.department}
              </span>
              <span class="flex items-center">
                <svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
                </svg>
                {job.location}
              </span>
              {#if job.employmentType}
                <span class={`badge ${getTypeBadgeClass(job.employmentType)}`}>
                  {formatEmploymentType(job.employmentType)}
                </span>
              {/if}
              {#if job.salaryRange}
                <span class="font-semibold text-gray-900">
                  {formatSalary(job.salaryRange)}
                </span>
              {/if}
            </div>
          </div>
          
          <button 
            on:click={() => showApplicationForm = !showApplicationForm} 
            class="btn btn-primary ml-4"
            disabled={submitSuccess}
          >
            {submitSuccess ? 'Application Submitted' : 'Apply Now'}
          </button>
        </div>
      </div>
    </div>

    <!-- Content -->
    <div class="max-w-5xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <div class="grid grid-cols-1 lg:grid-cols-3 gap-8">
        <!-- Main Content -->
        <div class="lg:col-span-2 space-y-8">
          <!-- Description -->
          <div class="card">
            <h2 class="text-2xl font-bold mb-4">About the Role</h2>
            <p class="text-gray-700 whitespace-pre-line">{job.description}</p>
          </div>

          <!-- Responsibilities -->
          {#if job.responsibilities}
            <div class="card">
              <h2 class="text-2xl font-bold mb-4">Responsibilities</h2>
              <div class="text-gray-700 whitespace-pre-line">{job.responsibilities}</div>
            </div>
          {/if}

          <!-- Requirements -->
          {#if job.requirements}
            <div class="card">
              <h2 class="text-2xl font-bold mb-4">Requirements</h2>
              <div class="text-gray-700 whitespace-pre-line">{job.requirements}</div>
            </div>
          {/if}

          <!-- Benefits -->
          {#if job.benefits}
            <div class="card">
              <h2 class="text-2xl font-bold mb-4">Benefits</h2>
              <div class="text-gray-700 whitespace-pre-line">{job.benefits}</div>
            </div>
          {/if}

          <!-- Skills -->
          {#if job.skills && job.skills.length > 0}
            <div class="card">
              <h2 class="text-2xl font-bold mb-4">Required Skills</h2>
              <div class="flex flex-wrap gap-2">
                {#each job.skills as skill}
                  <span class="badge badge-primary">{skill}</span>
                {/each}
              </div>
            </div>
          {/if}
        </div>

        <!-- Sidebar -->
        <div class="space-y-6">
          <!-- Job Stats -->
          <div class="card">
            <h3 class="text-lg font-semibold mb-4">Job Information</h3>
            <dl class="space-y-3 text-sm">
              {#if job.postedDate}
                <div>
                  <dt class="text-gray-600">Posted</dt>
                  <dd class="font-medium">{new Date(job.postedDate).toLocaleDateString()}</dd>
                </div>
              {/if}
              <div>
                <dt class="text-gray-600">Applications</dt>
                <dd class="font-medium">{job.applicationCount || 0}</dd>
              </div>
              <div>
                <dt class="text-gray-600">Views</dt>
                <dd class="font-medium">{job.viewCount || 0}</dd>
              </div>
              {#if job.experienceLevel}
                <div>
                  <dt class="text-gray-600">Experience Level</dt>
                  <dd class="font-medium">{formatEmploymentType(job.experienceLevel)}</dd>
                </div>
              {/if}
              {#if job.remoteWork !== undefined}
                <div>
                  <dt class="text-gray-600">Remote Work</dt>
                  <dd class="font-medium">{job.remoteWork ? 'Yes' : 'No'}</dd>
                </div>
              {/if}
            </dl>
          </div>

          <!-- Application CTA -->
          {#if !showApplicationForm && !submitSuccess}
            <div class="card bg-primary-50 border-primary-200">
              <h3 class="text-lg font-semibold mb-2">Interested?</h3>
              <p class="text-sm text-gray-600 mb-4">
                Apply now to join our team and make an impact.
              </p>
              <button on:click={() => showApplicationForm = true} class="btn btn-primary w-full">
                Apply for this Position
              </button>
            </div>
          {/if}
        </div>
      </div>

      <!-- Application Form -->
      {#if showApplicationForm && !submitSuccess}
        <div class="mt-8 card">
          <h2 class="text-2xl font-bold mb-6">Apply for {job.title}</h2>
          
          <form on:submit|preventDefault={handleSubmit} class="space-y-6">
            <!-- Personal Information -->
            <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div>
                <label for="firstName" class="label">First Name *</label>
                <input
                  id="firstName"
                  type="text"
                  bind:value={formData.firstName}
                  required
                  class="input"
                />
              </div>
              <div>
                <label for="lastName" class="label">Last Name *</label>
                <input
                  id="lastName"
                  type="text"
                  bind:value={formData.lastName}
                  required
                  class="input"
                />
              </div>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div>
                <label for="email" class="label">Email *</label>
                <input
                  id="email"
                  type="email"
                  bind:value={formData.email}
                  required
                  class="input"
                />
              </div>
              <div>
                <label for="phone" class="label">Phone *</label>
                <input
                  id="phone"
                  type="tel"
                  bind:value={formData.phone}
                  required
                  class="input"
                />
              </div>
            </div>

            <!-- Resume Upload -->
            <div>
              <label for="resume" class="label">Resume *</label>
              <input
                id="resume"
                type="file"
                on:change={handleFileSelect}
                accept=".pdf,.doc,.docx"
                required
                class="input"
              />
              <p class="text-sm text-gray-500 mt-1">PDF, DOC, or DOCX (Max 5MB)</p>
            </div>

            <!-- Cover Letter -->
            <div>
              <label for="coverLetter" class="label">Cover Letter</label>
              <textarea
                id="coverLetter"
                bind:value={formData.coverLetter}
                rows="6"
                class="input"
                placeholder="Tell us why you're interested in this position..."
              ></textarea>
            </div>

            <!-- URLs -->
            <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div>
                <label for="linkedinUrl" class="label">LinkedIn Profile</label>
                <input
                  id="linkedinUrl"
                  type="url"
                  bind:value={formData.linkedinUrl}
                  class="input"
                  placeholder="https://linkedin.com/in/..."
                />
              </div>
              <div>
                <label for="portfolioUrl" class="label">Portfolio / Website</label>
                <input
                  id="portfolioUrl"
                  type="url"
                  bind:value={formData.portfolioUrl}
                  class="input"
                  placeholder="https://..."
                />
              </div>
            </div>

            <!-- Additional Required Fields -->
            <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div>
                <label for="currentLocation" class="label">Current Location *</label>
                <input
                  id="currentLocation"
                  type="text"
                  bind:value={formData.currentLocation}
                  required
                  class="input"
                  placeholder="City, State"
                />
              </div>
              <div>
                <label for="availability" class="label">Availability *</label>
                <select
                  id="availability"
                  bind:value={formData.availability}
                  required
                  class="input"
                >
                  <option value="Immediate">Immediate</option>
                  <option value="2 weeks">2 Weeks Notice</option>
                  <option value="1 month">1 Month</option>
                  <option value="2-3 months">2-3 Months</option>
                </select>
              </div>
            </div>

            <!-- Error Message -->
            {#if submitError}
              <div class="bg-red-50 border border-red-200 rounded-lg p-4">
                <p class="text-red-800">{submitError}</p>
              </div>
            {/if}

            <!-- Submit Button -->
            <div class="flex justify-end space-x-4">
              <button
                type="button"
                on:click={() => showApplicationForm = false}
                class="btn btn-secondary"
                disabled={submitting}
              >
                Cancel
              </button>
              <button
                type="submit"
                class="btn btn-primary"
                disabled={submitting || resumeUploading}
              >
                {#if resumeUploading}
                  Uploading Resume...
                {:else if submitting}
                  Submitting...
                {:else}
                  Submit Application
                {/if}
              </button>
            </div>
          </form>
        </div>
      {/if}

      <!-- Success Message -->
      {#if submitSuccess}
        <div class="mt-8 card bg-green-50 border-green-200 text-center py-12">
          <svg class="w-16 h-16 text-green-600 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <h3 class="text-2xl font-bold text-green-900 mb-2">Application Submitted!</h3>
          <p class="text-green-700 mb-6">
            Thank you for applying to {job.title}. We'll review your application and get back to you soon.
          </p>
          <button on:click={() => navigate('/')} class="btn btn-primary">
            Back to Jobs
          </button>
        </div>
      {/if}
    </div>
  {/if}
</div>
