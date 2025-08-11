import {createFileRoute} from '@tanstack/react-router'
import {Heading} from "@/components/heading.tsx";
import Card from "@/components/card.tsx";
import {usePolicy} from "@/services/policies.ts";
import {ScaleLoader} from "react-spinners";
import {Textarea} from "@/components/textarea.tsx";
import {DescriptionDetails, DescriptionList, DescriptionTerm} from "@/components/description-list.tsx";
import {Button} from "@/components/button.tsx";

export const Route = createFileRoute('/policies/$language/$policyId/')({
    component: RouteComponent,
})

function RouteComponent() {
    const {language, policyId} = Route.useParams()
    const {status, data, error} = usePolicy(language, policyId)

    if (status == "pending") {
        return <>
            <div className="flex items-center justify-between">
                <Heading className="lg:text-3xl ">Policy</Heading>
            </div>
            <div className="flex flex-col 2xl:flex-row py-3 gap-6">
                <div className="mx-auto mt-4 flex w-[200px] items-center justify-center gap-y-2 flex-col">
                    <ScaleLoader height={16}/>
                    <div>Loading data...</div>
                </div>
            </div>
        </>
    }

    if (status == "error") {
        return <>
            <span>Error: {error.message}</span>
        </>
    }

    return <>
        <div className="flex items-center justify-between">
            <h1 className="text-rhc-color-cool-grey-900 font-normal text-[32px] leading-10">Beleid</h1>
        </div>
        <div className="flex flex-col 2xl:flex-row py-3 gap-6">
            <Card className="w-2/3 min-w-3xl flex-1 h-[836px]" header={
                <div className="flex items-center justify-between gap-4">
                    <Heading className="truncate">{data?.metadata?.title}</Heading>
                    <Button href={`/policies/${language}/${policyId}/edit`}>Edit</Button>
                </div>
            }>
                <div className="flex flex-col h-full bg-content-tertiary">
                    <Textarea name="data" value={data?.data} readOnly={true} className="flex-1"/>
                </div>
            </Card>
            <div className="w-1/3 min-w-3xl 2xl:min-w-lg flex flex-col gap-6">
                <Card className="flex-1" header={
                    <Heading>Metadata</Heading>
                }>
                    <DescriptionList>
                        <DescriptionTerm>Description</DescriptionTerm>
                        <DescriptionDetails>{data?.metadata?.description}</DescriptionDetails>
                    </DescriptionList>
                </Card>
                <Card className="flex-1" header={
                    <Heading>Versies</Heading>
                }>
                </Card>
            </div>
        </div>
    </>
}
