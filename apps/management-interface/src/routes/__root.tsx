import {Navbar} from '@/components/ui/navbar.tsx'
import {Sidebar, SidebarBody, SidebarItem, SidebarLabel, SidebarSection, SidebarSpacer,} from '@/components/ui/sidebar.tsx'
import {SidebarLayout} from '@/components/ui/sidebar-layout.tsx'
import {createRootRoute, Outlet} from '@tanstack/react-router'
import {TanStackRouterDevtools} from '@tanstack/react-router-devtools'
import {Header} from "@/components/ui/header.tsx";
import { menuItems } from "@/config/menu";
import { Suspense } from 'react';
import {useAuth} from "react-oidc-context";
import {useCapabilities} from "@/auth/useCapabilities";


const RootComponent = () => {
  const auth = useAuth()
  const { isAdmin } = useCapabilities()

  const labelClass = 'text-rhc-sidenav-link-font-size/7 font-rhc-sidenav-link-font-weight'

  const renderItem = (item: (typeof menuItems)[number]) => {
    const Icon = item.icon
    const disabled = item.disabled || (item.adminOnly && !isAdmin)
    return (
      <SidebarItem
        key={`${item.section}-${item.label}`}
        href={item.href}
        className={'text-rhc-sidenav-link-color '}
        disabled={disabled}
      >
        <Icon />
        <SidebarLabel className={labelClass}>{item.label}</SidebarLabel>
      </SidebarItem>
    )
  }

  const top = menuItems.filter(i => i.section === 'top')
  const bottom = menuItems.filter(i => i.section === 'bottom')

  function logoutHandler(): void {
    if (!auth.isAuthenticated || !auth.user) {
      return
    }

    void auth.signoutRedirect()
  }

  return (
    <>
      <Header onLogoutHandler={logoutHandler}/>
      <div className="px-3 xl:px-[150px]">
        <SidebarLayout
            navbar={
              <Navbar></Navbar>
            }
            sidebar={
              <Sidebar>
                <SidebarBody>
                  <SidebarSection className="gap-2">
                  {top.map(renderItem)}
                  <SidebarSpacer/>
                  {bottom.map(renderItem)}
                  </SidebarSection>
                </SidebarBody>
              </Sidebar>
            }
        >
          <Suspense fallback={<div className="p-4 text-sm text-content-secondary">Loading…</div>}>
            <Outlet/>
          </Suspense>
          <TanStackRouterDevtools/>
        </SidebarLayout>
      </div>
    </>
  );
};

export const Route = createRootRoute({
  component: RootComponent,
})
