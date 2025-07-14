import { Heading } from '@/components/heading'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/table'
// import { createFileRoute } from '@tanstack/react-router'
import {usePolicies} from "@/services/policies.ts";
import {ScaleLoader} from "react-spinners";
import {createFileRoute} from "@tanstack/react-router";
import {Button} from "@/components/button.tsx";

export const Route = createFileRoute('/policies/')({
  component: PoliciesComponent,
})

export default function PoliciesComponent() {
    const { data, isLoading, error } = usePolicies();

    if (isLoading) {
        return (
            <>
                <Heading>Policies</Heading>
                <div className="mx-auto mt-4 flex w-[200px] items-center justify-center gap-y-2 flex-col">
                    <ScaleLoader height={16}/>
                    <div>Loading data...</div>
                </div>
            </>
        )
    }
    if (error) {
        return <div>Error: {error.message}</div>;
    }

    return (
      <>
          <div className="flex items-center justify-between">
              <Heading>Policies</Heading>
              <Button color={"emerald"} href={"/policies/add"}>Add policy</Button>
          </div>
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
                  {data?.map((order) => (
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
