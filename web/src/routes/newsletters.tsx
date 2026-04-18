import { createFileRoute, Outlet } from '@tanstack/react-router'

export const Route = createFileRoute('/newsletters')({
  component: () => <Outlet />,
})
