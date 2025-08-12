import {Navbar} from '@/components/navbar'
import {Sidebar, SidebarBody, SidebarItem, SidebarLabel, SidebarSection, SidebarSpacer,} from '@/components/sidebar'
import {SidebarLayout} from '@/components/sidebar-layout'
import {createRootRoute, Outlet} from '@tanstack/react-router'
import {TanStackRouterDevtools} from '@tanstack/react-router-devtools'
import {
  IconDatabase,
  IconFileText,
  IconHelpCircleFilled,
  IconLayoutDashboard,
  IconLogs,
  IconNews,
  IconSettings
} from "@tabler/icons-react";
import {Header} from "@/components/header.tsx";


const RootComponent = () => {
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
                    <SidebarItem href="/" className={"text-rhc-sidenav-link-color"}>
                      <IconLayoutDashboard/>
                      <SidebarLabel className="text-rhc-sidenav-link-font-size/7 font-rhc-sidenav-link-font-weight">Overzicht</SidebarLabel>
                    </SidebarItem>
                    <SidebarItem href="/policies" className={"text-rhc-sidenav-link-color"}>
                      <IconFileText/>
                      <SidebarLabel className="text-rhc-sidenav-link-font-size/7 font-rhc-sidenav-link-font-weight">Beleid</SidebarLabel>
                    </SidebarItem>
                    <SidebarItem href="/attributen" className={"text-rhc-sidenav-link-color"}>
                      <IconDatabase/>
                      <SidebarLabel className="text-rhc-sidenav-link-font-size/7 font-rhc-sidenav-link-font-weight">Bronnen</SidebarLabel>
                    </SidebarItem>
                    <SidebarItem href="#" className={"text-rhc-sidenav-link-color"} disabled={true}>
                      <IconLogs/>
                      <SidebarLabel className="text-rhc-sidenav-link-font-size/7 font-rhc-sidenav-link-font-weight">Logboek</SidebarLabel>
                    </SidebarItem>
                    <SidebarItem href="#" className={"text-rhc-sidenav-link-color"} disabled={true}>
                      <IconSettings/>
                      <SidebarLabel className="text-rhc-sidenav-link-font-size/7 font-rhc-sidenav-link-font-weight">Systeem</SidebarLabel>
                    </SidebarItem>
                    <SidebarSpacer/>
                    <SidebarItem href="#" className={"text-rhc-sidenav-link-color"}>
                      <IconHelpCircleFilled/>
                      <SidebarLabel className="text-rhc-sidenav-link-font-size/7 font-rhc-sidenav-link-font-weight">Ondersteuning</SidebarLabel>
                    </SidebarItem>
                    <SidebarItem href="#" className={"text-rhc-sidenav-link-color"}>
                      <IconNews/>
                      <SidebarLabel className="text-rhc-sidenav-link-font-size/7 font-rhc-sidenav-link-font-weight">Nieuws</SidebarLabel>
                    </SidebarItem>
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
