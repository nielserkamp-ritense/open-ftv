import {Navbar} from '@/components/navbar'
import {Sidebar, SidebarBody, SidebarItem, SidebarLabel, SidebarSection, SidebarSpacer,} from '@/components/sidebar'
import {SidebarLayout} from '@/components/sidebar-layout'
import {createRootRoute, Outlet} from '@tanstack/react-router'
import {TanStackRouterDevtools} from '@tanstack/react-router-devtools'
import {Header} from "@/components/header.tsx";
import { menuItems } from "@/config/menu";


const RootComponent = () => {
  const linkClass = 'text-rhc-sidenav-link-color'
  const labelClass = 'text-rhc-sidenav-link-font-size/7 font-rhc-sidenav-link-font-weight'

  const renderItem = (item: (typeof menuItems)[number]) => {
    const Icon = item.icon
    return (
      <SidebarItem
        key={`${item.section}-${item.label}`}
        href={item.href}
        className={linkClass}
        disabled={item.disabled}
      >
        <Icon />
        <SidebarLabel className={labelClass}>{item.label}</SidebarLabel>
      </SidebarItem>
    )
  }

  const top = menuItems.filter(i => i.section === 'top')
  const bottom = menuItems.filter(i => i.section === 'bottom')

  return (
    <>
      <Header/>
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
          <Outlet/>
          <TanStackRouterDevtools/>
        </SidebarLayout>
      </div>
    </>
  );
};

export const Route = createRootRoute({
  component: RootComponent,
})
