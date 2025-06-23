import { Heading } from '@/components/heading'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/table'
import { GetPolicies } from './../data/policy'
import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/policies')({
  component: PoliciesComponent,
})

export default function PoliciesComponent() {  
  return (
      <>
          <Heading>Policies</Heading>
          <Table className="mt-4 [--gutter:--spacing(6)] lg:[--gutter:--spacing(10)]">
              <TableHead>
                  <TableRow>
                      <TableHeader>ID</TableHeader>
                      <TableHeader>Language</TableHeader>
                      <TableHeader>URL</TableHeader>
                      <TableHeader className="text-right">rvvaId</TableHeader>
                  </TableRow>
              </TableHead>
              <TableBody>
                  {GetPolicies().map((order) => (
                      <TableRow key={order.id} href={order.url} title={`Policy #${order.id}`}>
                          <TableCell>{order.id}</TableCell>
                          <TableCell className="text-zinc-500">{order.language}</TableCell>
                          <TableCell>{order.url}</TableCell>
                          <TableCell className="text-right">{order.rvvaId}</TableCell>
                      </TableRow>
                  ))}
              </TableBody>
          </Table>
      </>
  )
}
