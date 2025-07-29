import { Avatar } from '@/components/avatar'
import {
  Dropdown,
  DropdownButton,
  DropdownItem,
  DropdownLabel,
  DropdownMenu,
} from '@/components/dropdown'
import { Navbar } from '@/components/navbar'
import {
  Sidebar,
  SidebarBody,
  SidebarHeader,
  SidebarItem,
  SidebarLabel,
  SidebarSection,
  SidebarSpacer,
} from '@/components/sidebar'
import { SidebarLayout } from '@/components/sidebar-layout'
import {
  ChevronDownIcon,
  Cog8ToothIcon,
} from '@heroicons/react/16/solid'
import {
  HomeIcon,
  QuestionMarkCircleIcon,
  SparklesIcon,
} from '@heroicons/react/20/solid'
import { Square3Stack3DIcon, AdjustmentsHorizontalIcon } from '@heroicons/react/24/outline'
import { createRootRoute, Outlet } from '@tanstack/react-router'
import { TanStackRouterDevtools } from '@tanstack/react-router-devtools'


export const Route = createRootRoute({
  component: () => (
    <>
      <div className="px-3 xl:px-[150px]">
        <SidebarLayout
            navbar={
              <Navbar></Navbar>
            }
            sidebar={
              <Sidebar>
                <SidebarHeader>
                  <Dropdown>
                    <DropdownButton as={SidebarItem} className="lg:mb-2.5">
                      <Avatar src="/logo-ftv.png" square />
                      <SidebarLabel>FTV Portal</SidebarLabel>
                      <ChevronDownIcon />
                    </DropdownButton>
                    <DropdownMenu className="min-w-80 lg:min-w-64" anchor="bottom start">
                      <DropdownItem href="/teams/1/settings">
                        <Cog8ToothIcon />
                        <DropdownLabel>Settings</DropdownLabel>
                      </DropdownItem>
                    </DropdownMenu>
                  </Dropdown>
                </SidebarHeader>
                <SidebarBody>
                  <SidebarSection>
                    <SidebarItem href="/">
                      <HomeIcon />
                      <SidebarLabel>Home</SidebarLabel>
                    </SidebarItem>
                    <SidebarItem href="/policies">
                      <Square3Stack3DIcon />
                      <SidebarLabel>Policies</SidebarLabel>
                    </SidebarItem>
                    <SidebarItem href="/attributes">
                      <AdjustmentsHorizontalIcon />
                      <SidebarLabel>Attributes</SidebarLabel>
                    </SidebarItem>
                  </SidebarSection>
                  <SidebarSpacer />
                  <SidebarSection>
                    <SidebarItem href="https://vng-realisatie.github.io/ftv/">
                      <QuestionMarkCircleIcon />
                      <SidebarLabel>Mattermost</SidebarLabel>
                    </SidebarItem>
                    <SidebarItem href="/changelog">
                      <SparklesIcon />
                      <SidebarLabel>Changelog</SidebarLabel>
                    </SidebarItem>
                  </SidebarSection>
                </SidebarBody>
              </Sidebar>
            }
        >
          <Outlet />
          <TanStackRouterDevtools />
        </SidebarLayout>
      </div>
    </>
  ),
})

