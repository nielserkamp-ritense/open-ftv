import { createFileRoute, Link } from '@tanstack/react-router'
import { Square3Stack3DIcon, Cog8ToothIcon, AdjustmentsHorizontalIcon } from '@heroicons/react/24/outline'

const menuItems = 5

export const Route = createFileRoute('/')({
  component: RouteComponent,
})

function RouteComponent() {
  return <>
    <div className="grid grid-cols-4 grid-rows-2 gap-4 sm:grid-cols-4">
        <div
          key="Policies"
          className="relative flex items-center space-x-3 rounded-sm border border-gray-300 bg-white px-6 py-5 shadow-md focus-within:ring-2 focus-within:ring-indigo-500 focus-within:ring-offset-2 hover:border-gray-400"
        >
          <div className="shrink-0">
            <Square3Stack3DIcon className="size-16 rounded-full" />
          </div>
          <div className="min-w-0 flex-1">
            <Link to="/policies" className="focus:outline-hidden">
              <span aria-hidden="true" className="absolute inset-0" />
              <p className="text-sm font-medium text-gray-900">Policies</p>
            </Link>
          </div>
        </div>
        <div
          key="attributes"
          className="relative flex items-center space-x-3 rounded-sm border border-gray-300 bg-white px-6 py-5 shadow-md focus-within:ring-2 focus-within:ring-indigo-500 focus-within:ring-offset-2 hover:border-gray-400"
        >
          <div className="shrink-0">
            <AdjustmentsHorizontalIcon className="size-16 rounded-full" />
          </div>
          <div className="min-w-0 flex-1">
            <Link to="/attributen" className="focus:outline-hidden">
              <span aria-hidden="true" className="absolute inset-0" />
              <p className="text-sm font-medium text-gray-900">Attributes</p>
            </Link>
          </div>
        </div>
        <div
          key="settings"
          className="relative flex items-center space-x-3 rounded-sm border border-gray-300 bg-white px-6 py-5 shadow-md focus-within:ring-2 focus-within:ring-indigo-500 focus-within:ring-offset-2 hover:border-gray-400"
        >
          <div className="shrink-0">
            <Cog8ToothIcon className="size-16 rounded-full" />
          </div>
          <div className="min-w-0 flex-1">
            <a href="#" className="focus:outline-hidden">
              <span aria-hidden="true" className="absolute inset-0" />
              <p className="text-sm font-medium text-gray-900">Settings</p>
            </a>
          </div>
        </div>
        
        {Array.from({ length: menuItems }).map((_, index) => (
          <div
            key={`menu-item-${index}`}
            className="relative flex items-center space-x-3 rounded-sm border border-gray-300 px-6 py-5 bg-gray-100 shadow-md"
          >
            <div className="shrink-0">
              
            </div>
            <div className="min-w-0 flex-1">
            </div>
          </div>
        ))}
    </div>
  </>
}
