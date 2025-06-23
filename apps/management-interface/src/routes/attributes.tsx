import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/attributes')({
  component: RouteComponent,
})

function RouteComponent() {
  return <div>Hello "/attributes"!</div>
}
