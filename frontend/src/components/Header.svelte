<script lang="ts">
  function handleNavClick(e: MouseEvent, path: string) {
    e.preventDefault();
    if (typeof window !== 'undefined' && (window as any).navigate) {
      (window as any).navigate(path);
    }
  }
</script>

<header class="bg-white border-b border-gray-200 sticky top-0 z-50 shadow-sm">
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
    <div class="flex justify-between items-center h-16">
      <div class="flex items-center">
        <a href="/" on:click={(e) => handleNavClick(e, '/')} class="flex items-center space-x-3">
          <div class="w-10 h-10 bg-gradient-to-br from-primary-500 to-primary-700 rounded-lg flex items-center justify-center">
            <span class="text-white font-bold text-xl">C</span>
          </div>
          <div class="flex flex-col">
            <span class="text-xl font-bold text-gray-900">Company</span>
            <span class="text-xs text-gray-500">Careers</span>
          </div>
        </a>
      </div>

      <nav class="hidden md:flex space-x-8">
        <a 
          href="/" 
          on:click={(e) => handleNavClick(e, '/')}
          class="text-gray-700 hover:text-primary-600 px-3 py-2 text-sm font-medium transition-colors"
        >
          Browse Jobs
        </a>
        <a href="#about" class="text-gray-700 hover:text-primary-600 px-3 py-2 text-sm font-medium transition-colors">
          About Us
        </a>
        <a href="#culture" class="text-gray-700 hover:text-primary-600 px-3 py-2 text-sm font-medium transition-colors">
          Culture
        </a>
        <a href="#benefits" class="text-gray-700 hover:text-primary-600 px-3 py-2 text-sm font-medium transition-colors">
          Benefits
        </a>
      </nav>

      <div class="flex items-center space-x-4">
        <button class="btn btn-ghost md:hidden">
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
          </svg>
        </button>
      </div>
    </div>
  </div>
</header>
