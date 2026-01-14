<script lang="ts">
  import { Router, Route } from 'svelte-routing';
  import Header from './components/Header.svelte';
  import Footer from './components/Footer.svelte';
  import Home from './pages/Home.svelte';
  import JobDetail from './pages/JobDetail.svelte';
</script>

<Router>
  <div class="flex flex-col min-h-screen">
    <Header />
    
    <main class="flex-1">
      <Route path="/" component={Home} />
      <Route path="/jobs/:id" let:params>
        <JobDetail id={params.id} />
      </Route>
    </main>

    <Footer />
  </div>
</Router>
