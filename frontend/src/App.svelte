<script lang="ts">
  import { writable } from 'svelte/store';
  import Header from './components/Header.svelte';
  import Footer from './components/Footer.svelte';
  import Home from './pages/Home.svelte';
  import JobDetail from './pages/JobDetail.svelte';

  // Simple routing state
  const route = writable({ path: '/', params: {} });

  // Parse current URL
  function updateRoute() {
    const path = window.location.pathname;
    const jobMatch = path.match(/^\/jobs\/(.+)$/);
    
    if (jobMatch) {
      route.set({ path: 'job-detail', params: { id: jobMatch[1] } });
    } else {
      route.set({ path: 'home', params: {} });
    }
  }

  // Handle navigation
  function navigate(path: string) {
    window.history.pushState({}, '', path);
    updateRoute();
  }

  // Update route on page load and browser back/forward
  $: if (typeof window !== 'undefined') {
    updateRoute();
    window.addEventListener('popstate', updateRoute);
  }

  // Make navigate available globally
  if (typeof window !== 'undefined') {
    (window as any).navigate = navigate;
  }
</script>

<div class="flex flex-col min-h-screen">
  <Header />
  
  <main class="flex-1">
    {#if $route.path === 'home'}
      <Home />
    {:else if $route.path === 'job-detail'}
      <JobDetail id={$route.params.id} />
    {/if}
  </main>

  <Footer />
</div>
