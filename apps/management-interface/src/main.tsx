import { StrictMode, useEffect } from 'react'
import type { ReactNode } from 'react'
import ReactDOM from 'react-dom/client'
import { AuthProvider, useAuth } from 'react-oidc-context'
import { oidcConfig } from './auth/oidc'
import { setAccessToken } from './api/client/token'
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

function TokenSync() {
  const auth = useAuth()
  useEffect(() => { setAccessToken(auth.user?.access_token) }, [auth.user?.access_token])
  return null
}

function AuthGate({ children }: { children: ReactNode }) {
  const auth = useAuth()
  useEffect(() => {
    if (!auth.isLoading && !auth.isAuthenticated && !auth.activeNavigator && !auth.error) {
      void auth.signinRedirect()
    }
  }, [auth.isLoading, auth.isAuthenticated, auth.activeNavigator, auth.error])
  if (auth.error) return <div className="p-6">Authenticatie-fout: {auth.error.message}</div>
  if (auth.isLoading || !auth.isAuthenticated) return <div className="p-6">Bezig met inloggen…</div>
  return <>{children}</>
}

const rootElement = document.getElementById('root')!
if (!rootElement.innerHTML) {
  const root = ReactDOM.createRoot(rootElement)
  root.render(
    <AuthProvider {...oidcConfig}>
      <QueryClientProvider client={queryClient}>
        <StrictMode>
          <TokenSync />
          <AuthGate>
            <RouterProvider router={router} />
          </AuthGate>
        </StrictMode>
        <ReactQueryDevtools initialIsOpen={false} />
      </QueryClientProvider>
    </AuthProvider>
  )
}
