import { StrictMode } from 'react'
import ReactDOM from 'react-dom/client'
import { RouterProvider, createRouter } from '@tanstack/react-router'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ReactQueryDevtools } from '@tanstack/react-query-devtools';
import { AllCommunityModule, ModuleRegistry } from 'ag-grid-community';
import './index.css'

// Import the generated route tree
import { routeTree } from './routeTree.gen'
// Create a new router instance
const router = createRouter({ routeTree })

// Register the router instance for type safety
declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router
  }
}

// Register all Community features
ModuleRegistry.registerModules([AllCommunityModule]);

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 5 * 60 * 1000, // 5 minutes
      retry: 1,
    },
  },
});

const rootElement = document.getElementById('root')!
if (!rootElement.innerHTML) {
  const root = ReactDOM.createRoot(rootElement)
  root.render(
      <QueryClientProvider client={queryClient}>
        <StrictMode>
          <RouterProvider router={router} />
        </StrictMode>
        <ReactQueryDevtools initialIsOpen={false} />
      </QueryClientProvider>
  )
}
