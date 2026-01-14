<script lang="ts">
  import { onMount } from 'svelte';
  import { navigate } from 'svelte-routing';
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
      const resumeUrl = await uploadAPI.uploadFile(formData.resumeFile);
      resumeUploading = false;

      // Submit application
      await applicationsAPI.submit({
        job_id: id,
        first_name: formData.firstName,
        last_name: formData.lastName,
        email: formData.email,
        phone: formData.phone,
        resume_url: resumeUrl,
        cover_letter: formData.coverLetter || undefined,
        linkedin_url: formData.linkedinUrl || undefined,
        portfolio_url: formData.portfolioUrl || undefined,
      });

      submitSuccess = true;
    } catch (err) {
      submitError = err instanceof Error ? err.message : 'Failed to submit application';
      submitting = false;
    }
  }

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

  const typeColors = {
    'full-time': 'badge-primary',
    'part-time': 'badge-success',
    'contract': 'badge-warning',
    'internship': 'badge-gray',
  };
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
              <span class={`badge ${typeColors[job.type]}`}>
                {job.type.replace('-', ' ')}
              </span>
              {#if job.salary_range}
                <span class="font-semibold text-gray-900">
                  {formatSalary(job.salary_range)}
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
          {#if job.responsibilities && job.responsibilities.length > 0}
            <div class="card">
              <h2 class="text-2xl font-bold mb-4">Responsibilities</h2>
              <ul class="space-y-2">
                {#each job.responsibilities as responsibility}
                  <li class="flex items-start">
                    <svg class="w-5 h-5 text-primary-600 mr-3 mt-0.5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                    <span class="text-gray-700">{responsibility}</span>
                  </li>
                {/each}
              </ul>
            </div>
          {/if}

          <!-- Requirements -->
          {#if job.requirements && job.requirements.length > 0}
            <div class="card">
              <h2 class="text-2xl font-bold mb-4">Requirements</h2>
              <ul class="space-y-2">
                {#each job.requirements as requirement}
                  <li class="flex items-start">
                    <svg class="w-5 h-5 text-primary-600 mr-3 mt-0.5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                    <span class="text-gray-700">{requirement}</span>
                  </li>
                {/each}
              </ul>
            </div>
          {/if}

          <!-- Benefits -->
          {#if job.benefits && job.benefits.length > 0}
            <div class="card">
              <h2 class="text-2xl font-bold mb-4">Benefits</h2>
              <ul class="grid grid-cols-1 md:grid-cols-2 gap-3">
                {#each job.benefits as benefit}
                  <li class="flex items-center">
                    <svg class="w-5 h-5 text-green-600 mr-3 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                    </svg>
                    <span class="text-gray-700">{benefit}</span>
                  </li>
                {/each}
              </ul>
            </div>
          {/if}
        </div>

        <!-- Sidebar / Application Form -->
        <div class="lg:col-span-1">
          {#if !showApplicationForm && !submitSuccess}
            <div class="card sticky top-20">
              <h3 class="text-xl font-bold mb-4">Ready to Apply?</h3>
              <p class="text-gray-600 mb-6">
                Join our team and make an impact. Click below to start your application.
              </p>
              <button on:click={() => showApplicationForm = true} class="btn btn-primary w-full">
                Apply for this Position
              </button>
            </div>
          {:else if submitSuccess}
            <div class="card sticky top-20 bg-green-50 border-green-200">
              <div class="text-center">
                <svg class="w-16 h-16 text-green-500 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                <h3 class="text-xl font-bold text-green-900 mb-2">Application Submitted!</h3>
                <p class="text-green-700 mb-6">
                  Thank you for applying. We'll review your application and get back to you soon.
                </p>
                <button on:click={() => navigate('/')} class="btn btn-secondary w-full">
                  Browse More Jobs
                </button>
              </div>
            </div>
          {:else}
            <div class="card sticky top-20">
              <h3 class="text-xl font-bold mb-6">Apply for {job.title}</h3>
              
              <form on:submit|preventDefault={handleSubmit} class="space-y-4">
                <!-- Name -->
                <div class="grid grid-cols-2 gap-4">
                  <div>
                    <label for="firstName" class="label">First Name *</label>
                    <input
                      id="firstName"
                      type="text"
                      bind:value={formData.firstName}
                      required
                      class="input"
                      disabled={submitting}
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
                      disabled={submitting}
                    />
                  </div>
                </div>

                <!-- Email -->
                <div>
                  <label for="email" class="label">Email *</label>
                  <input
                    id="email"
                    type="email"
                    bind:value={formData.email}
                    required
                    class="input"
                    disabled={submitting}
                  />
                </div>

                <!-- Phone -->
                <div>
                  <label for="phone" class="label">Phone *</label>
                  <input
                    id="phone"
                    type="tel"
                    bind:value={formData.phone}
                    required
                    class="input"
                    disabled={submitting}
                  />
                </div>

                <!-- Resume -->
                <div>
                  <label for="resume" class="label">Resume * (PDF, DOC, DOCX)</label>
                  <input
                    id="resume"
                    type="file"
                    accept=".pdf,.doc,.docx"
                    on:change={handleFileSelect}
                    required
                    class="input"
                    disabled={submitting}
                  />
                  {#if formData.resumeFile}
                    <p class="text-sm text-gray-600 mt-1">
                      {formData.resumeFile.name}
                    </p>
                  {/if}
                </div>

                <!-- Cover Letter -->
                <div>
                  <label for="coverLetter" class="label">Cover Letter</label>
                  <textarea
                    id="coverLetter"
                    bind:value={formData.coverLetter}
                    rows="4"
                    class="textarea"
                    disabled={submitting}
                    placeholder="Tell us why you're a great fit..."
                  ></textarea>
                </div>

                <!-- LinkedIn -->
                <div>
                  <label for="linkedin" class="label">LinkedIn URL</label>
                  <input
                    id="linkedin"
                    type="url"
                    bind:value={formData.linkedinUrl}
                    class="input"
                    disabled={submitting}
                    placeholder="https://linkedin.com/in/yourprofile"
                  />
                </div>

                <!-- Portfolio -->
                <div>
                  <label for="portfolio" class="label">Portfolio URL</label>
                  <input
                    id="portfolio"
                    type="url"
                    bind:value={formData.portfolioUrl}
                    class="input"
                    disabled={submitting}
                    placeholder="https://yourportfolio.com"
                  />
                </div>

                <!-- Error Message -->
                {#if submitError}
                  <div class="bg-red-50 border border-red-200 rounded-lg p-3 text-sm text-red-700">
                    {submitError}
                  </div>
                {/if}

                <!-- Submit Button -->
                <button type="submit" class="btn btn-primary w-full" disabled={submitting}>
                  {#if resumeUploading}
                    Uploading Resume...
                  {:else if submitting}
                    Submitting Application...
                  {:else}
                    Submit Application
                  {/if}
                </button>

                <button 
                  type="button" 
                  on:click={() => showApplicationForm = false} 
                  class="btn btn-ghost w-full"
                  disabled={submitting}
                >
                  Cancel
                </button>
              </form>
            </div>
          {/if}
        </div>
      </div>
    </div>
  {/if}
</div>
