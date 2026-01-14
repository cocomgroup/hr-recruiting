<script lang="ts">
  const currentYear = new Date().getFullYear();
</script>

<footer class="bg-gray-900 text-gray-300">
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
    <div class="grid grid-cols-1 md:grid-cols-4 gap-8">
      <!-- Company Info -->
      <div class="col-span-1">
        <div class="flex items-center space-x-2 mb-4">
          <div class="w-8 h-8 bg-gradient-to-br from-primary-500 to-primary-700 rounded-lg flex items-center justify-center">
            <span class="text-white font-bold">C</span>
          </div>
          <span class="text-white font-bold text-lg">Company</span>
        </div>
        <p class="text-sm text-gray-400">
          Building the future together. Join our team and make an impact.
        </p>
        <div class="flex space-x-4 mt-4">
          <a href="#" class="text-gray-400 hover:text-white transition-colors">
            <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
              <path d="M22 12c0-5.523-4.477-10-10-10S2 6.477 2 12c0 4.991 3.657 9.128 8.438 9.878v-6.987h-2.54V12h2.54V9.797c0-2.506 1.492-3.89 3.777-3.89 1.094 0 2.238.195 2.238.195v2.46h-1.26c-1.243 0-1.63.771-1.63 1.562V12h2.773l-.443 2.89h-2.33v6.988C18.343 21.128 22 16.991 22 12z"/>
            </svg>
          </a>
          <a href="#" class="text-gray-400 hover:text-white transition-colors">
            <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
              <path d="M23 3a10.9 10.9 0 01-3.14 1.53 4.48 4.48 0 00-7.86 3v1A10.66 10.66 0 013 4s-4 9 5 13a11.64 11.64 0 01-7 2c9 5 20 0 20-11.5a4.5 4.5 0 00-.08-.83A7.72 7.72 0 0023 3z"/>
            </svg>
          </a>
          <a href="#" class="text-gray-400 hover:text-white transition-colors">
            <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
              <path d="M16 8a6 6 0 016 6v7h-4v-7a2 2 0 00-2-2 2 2 0 00-2 2v7h-4v-7a6 6 0 016-6zM2 9h4v12H2z"/>
              <circle cx="4" cy="4" r="2"/>
            </svg>
          </a>
        </div>
      </div>

      <!-- Quick Links -->
      <div>
        <h3 class="text-white font-semibold mb-4">Company</h3>
        <ul class="space-y-2">
          <li><a href="#" class="text-sm hover:text-white transition-colors">About Us</a></li>
          <li><a href="#" class="text-sm hover:text-white transition-colors">Our Team</a></li>
          <li><a href="#" class="text-sm hover:text-white transition-colors">Careers</a></li>
          <li><a href="#" class="text-sm hover:text-white transition-colors">Press</a></li>
        </ul>
      </div>

      <!-- Resources -->
      <div>
        <h3 class="text-white font-semibold mb-4">Resources</h3>
        <ul class="space-y-2">
          <li><a href="#" class="text-sm hover:text-white transition-colors">Blog</a></li>
          <li><a href="#" class="text-sm hover:text-white transition-colors">Help Center</a></li>
          <li><a href="#" class="text-sm hover:text-white transition-colors">Contact</a></li>
          <li><a href="#" class="text-sm hover:text-white transition-colors">FAQ</a></li>
        </ul>
      </div>

      <!-- Legal -->
      <div>
        <h3 class="text-white font-semibold mb-4">Legal</h3>
        <ul class="space-y-2">
          <li><a href="#" class="text-sm hover:text-white transition-colors">Privacy Policy</a></li>
          <li><a href="#" class="text-sm hover:text-white transition-colors">Terms of Service</a></li>
          <li><a href="#" class="text-sm hover:text-white transition-colors">Cookie Policy</a></li>
          <li><a href="#" class="text-sm hover:text-white transition-colors">GDPR</a></li>
        </ul>
      </div>
    </div>

    <div class="border-t border-gray-800 mt-8 pt-8 text-center">
      <p class="text-sm text-gray-400">
        © {currentYear} Company. All rights reserved. Equal Opportunity Employer.
      </p>
    </div>
  </div>
</footer>
