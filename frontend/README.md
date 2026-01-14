# HR-Recruiting Frontend

A modern, responsive career portal built with Svelte, TypeScript, and Tailwind CSS.

## Features

✨ **Modern UI/UX**
- Clean, professional design
- Fully responsive (mobile, tablet, desktop)
- Smooth animations and transitions
- Accessible components

🔍 **Job Search & Filtering**
- Real-time search
- Filter by department, location, and job type
- Advanced search capabilities

📝 **Application Process**
- Intuitive application form
- Resume upload with S3 integration
- Cover letter and portfolio support
- Real-time validation

🎨 **Design System**
- Tailwind CSS for styling
- Custom component library
- Consistent color palette
- Reusable design tokens

## Tech Stack

- **Framework**: Svelte 4
- **Language**: TypeScript
- **Styling**: Tailwind CSS
- **Routing**: svelte-routing
- **Build Tool**: Vite
- **Icons**: Heroicons / Lucide

## Getting Started

### Prerequisites

- Node.js 18+ and npm/yarn
- Backend API running on `http://localhost:8080`

### Installation

```bash
# Install dependencies
npm install

# Create environment file
cp .env.example .env

# Start development server
npm run dev
```

The application will be available at `http://localhost:5173`

### Build for Production

```bash
# Build optimized bundle
npm run build

# Preview production build
npm run preview
```

## Project Structure

```
src/
├── components/          # Reusable UI components
│   ├── Header.svelte   # Navigation header
│   ├── Footer.svelte   # Site footer
│   ├── JobCard.svelte  # Job listing card
│   └── JobFilters.svelte # Search and filter UI
├── pages/              # Page components
│   ├── Home.svelte     # Jobs listing page
│   └── JobDetail.svelte # Job details & application
├── stores/             # Svelte stores (state management)
│   └── jobs.ts         # Jobs state and filters
├── lib/                # Utilities and services
│   └── api.ts          # API client
├── App.svelte          # Root component
├── main.ts             # Entry point
└── app.css             # Global styles
```

## Configuration

### Environment Variables

Create a `.env` file:

```bash
VITE_API_URL=http://localhost:8080/api/v1
```

### API Configuration

The frontend communicates with the backend API through the `/api` proxy in development.

Edit `vite.config.ts` to change the backend URL:

```typescript
server: {
  proxy: {
    '/api': {
      target: 'http://your-backend:8080',
      changeOrigin: true,
    },
  },
}
```

## Key Components

### JobCard

Displays job summary in a card layout.

```svelte
<JobCard {job} />
```

### JobFilters

Search and filter interface.

```svelte
<JobFilters />
```

### JobDetail

Full job details with application form.

```svelte
<JobDetail id={jobId} />
```

## State Management

Uses Svelte stores for global state:

```typescript
// jobs.ts
export const jobsStore = createJobsStore();
export const filteredJobs = derived(jobsStore, ...);
```

### Using Stores

```svelte
<script>
  import { jobsStore, filteredJobs } from './stores/jobs';
  
  onMount(() => {
    jobsStore.fetchJobs();
  });
</script>

{#each $filteredJobs as job}
  <JobCard {job} />
{/each}
```

## API Integration

### API Client

The `api.ts` module provides type-safe API methods:

```typescript
import { jobsAPI, applicationsAPI, uploadAPI } from './lib/api';

// Fetch jobs
const jobs = await jobsAPI.list({ department: 'Engineering' });

// Get job details
const job = await jobsAPI.get('job-id');

// Submit application
await applicationsAPI.submit({
  job_id: 'job-id',
  first_name: 'John',
  last_name: 'Doe',
  // ...
});

// Upload resume
const url = await uploadAPI.uploadFile(file);
```

### Error Handling

API calls include comprehensive error handling:

```typescript
try {
  const jobs = await jobsAPI.list();
} catch (error) {
  if (error instanceof APIError) {
    console.error('API Error:', error.status, error.message);
  }
}
```

## Styling

### Tailwind Utilities

Custom utilities defined in `app.css`:

```css
/* Buttons */
.btn              /* Base button */
.btn-primary      /* Primary action */
.btn-secondary    /* Secondary action */
.btn-ghost        /* Minimal button */

/* Inputs */
.input            /* Text input */
.textarea         /* Text area */
.label            /* Form label */

/* Components */
.card             /* Card container */
.badge            /* Badge/tag */
```

### Using Components

```svelte
<button class="btn btn-primary">
  Apply Now
</button>

<div class="card">
  <h2 class="text-2xl font-bold">Title</h2>
  <p class="text-gray-600">Content</p>
</div>
```

## Customization

### Branding

Update colors in `tailwind.config.js`:

```javascript
theme: {
  extend: {
    colors: {
      primary: {
        50: '#eff6ff',
        // ... your brand colors
        900: '#1e3a8a',
      },
    },
  },
}
```

### Logo

Replace the logo in `Header.svelte`:

```svelte
<div class="w-10 h-10 ...">
  <!-- Your logo here -->
</div>
```

### Content

Update text and images throughout components to match your company.

## Performance

### Optimizations

- **Code Splitting**: Routes are automatically split
- **Tree Shaking**: Unused code is removed
- **Asset Optimization**: Images and CSS are optimized
- **Lazy Loading**: Components load on demand

### Bundle Size

```bash
npm run build
```

Typical bundle sizes:
- Main JS: ~50KB (gzipped)
- CSS: ~10KB (gzipped)
- Vendor: ~25KB (gzipped)

## Deployment

### Static Hosting

Deploy the `dist` folder to any static host:

**Netlify:**
```bash
# Build command
npm run build

# Publish directory
dist
```

**Vercel:**
```bash
vercel --prod
```

**AWS S3 + CloudFront:**
```bash
aws s3 sync dist/ s3://your-bucket/
aws cloudfront create-invalidation --distribution-id XXX --paths "/*"
```

### Docker

```dockerfile
FROM node:18-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

### Nginx Configuration

```nginx
server {
    listen 80;
    server_name _;
    root /usr/share/nginx/html;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location /api {
        proxy_pass http://backend:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

## Testing

### Type Checking

```bash
npm run check
```

### Manual Testing Checklist

- [ ] Jobs list loads correctly
- [ ] Search and filters work
- [ ] Job details page displays
- [ ] Application form validates
- [ ] Resume upload works
- [ ] Form submission succeeds
- [ ] Success message shows
- [ ] Mobile responsive
- [ ] Back button works

## Troubleshooting

### Common Issues

**CORS Errors**
- Ensure backend CORS allows your origin
- Check `CORS_ALLOWED_ORIGINS` in backend

**API Connection Failed**
- Verify backend is running
- Check `VITE_API_URL` in `.env`
- Verify proxy configuration in `vite.config.ts`

**Resume Upload Fails**
- Check S3 bucket permissions
- Verify presigned URL generation
- Check file size limits

**Build Errors**
- Clear node_modules: `rm -rf node_modules && npm install`
- Clear cache: `rm -rf .vite`
- Update dependencies: `npm update`

## Browser Support

- Chrome/Edge (last 2 versions)
- Firefox (last 2 versions)
- Safari (last 2 versions)
- iOS Safari (last 2 versions)
- Android Chrome (last 2 versions)

## Accessibility

- Semantic HTML
- ARIA labels where needed
- Keyboard navigation
- Screen reader friendly
- Color contrast compliance

## License

Private - Company Internal Use

## Support

For issues or questions, contact the development team.
