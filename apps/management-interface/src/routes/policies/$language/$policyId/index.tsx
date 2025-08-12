import {createFileRoute} from '@tanstack/react-router'
import {FTVHeading, Heading} from "@/components/heading.tsx";
import Card from "@/components/card.tsx";
import {usePolicy} from "@/services/policies.ts";
import {ScaleLoader} from "react-spinners";
import {Textarea} from "@/components/textarea.tsx";
import {DescriptionDetails, DescriptionList, DescriptionTerm} from "@/components/description-list.tsx";
import {Navbar, NavbarItem, NavbarSection} from "@/components/navbar.tsx";
import {IconFileText, IconPlayerPlay, IconSettings, IconSourceCode} from "@tabler/icons-react";

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
            <Card className="w-2/3 min-w-3xl flex-1 h-[836px]" disablePadding={true} header={
                <>
                    <Navbar className={"px-4 pt-5 sm:px-6"}>
                        <NavbarSection>
                            <NavbarItem href="#" className={"text-xl"} current>
                                <IconFileText />Details
                            </NavbarItem>
                            <NavbarItem href={"/policies/" + language +"/" + policyId + "/edit/" }>
                                <IconSourceCode /> Bewerken
                            </NavbarItem>
                            <NavbarItem href="#" disabled>
                                <IconPlayerPlay /> Test cases
                            </NavbarItem>
                            <NavbarItem href="#" disabled>
                                <IconSettings /> Instellingen
                            </NavbarItem>
                        </NavbarSection>
                    </Navbar>
                </>
            }>
                <div className="flex flex-col h-full px-4 py-5 sm:p-6">
                    <FTVHeading level={2}>{data?.metadata?.title}</FTVHeading>
                    <Textarea name="data" value={data?.data} readOnly={true} disabled={true} className="flex-1 bg-content-tertiary mt-3"/>
                </div>
            </Card>
            <div className="w-1/3 min-w-3xl 2xl:min-w-lg flex flex-col gap-6">
                <Card className="flex-1">
                    <DescriptionList>
                        <DescriptionTerm>Titel</DescriptionTerm>
                        <DescriptionDetails>{data?.metadata?.title}</DescriptionDetails>
                        <DescriptionTerm>Beschrijving</DescriptionTerm>
                        <DescriptionDetails>{data?.metadata?.description}</DescriptionDetails>
                        <DescriptionTerm>Taal</DescriptionTerm>
                        <DescriptionDetails>{data?.language?.charAt(0).toUpperCase() + data?.language?.slice(1).toLowerCase()}</DescriptionDetails>
                        <DescriptionTerm>rvvaID</DescriptionTerm>
                        <DescriptionDetails>{data?.metadata?.rvvaId}</DescriptionDetails>
                        <DescriptionTerm>URL</DescriptionTerm>
                        <DescriptionDetails>{data?.metadata?.url}</DescriptionDetails>
                        <DescriptionTerm>Tags</DescriptionTerm>
                        <DescriptionDetails>{data?.metadata?.tags}</DescriptionDetails>
                        <DescriptionTerm>Gemaakt door</DescriptionTerm>
                        <DescriptionDetails>{data?.metadata?.createUser}</DescriptionDetails>
                        <DescriptionTerm>Laatst bijgewerkt</DescriptionTerm>
                        <DescriptionDetails>{data?.metadata?.updateDate}</DescriptionDetails>
                        <DescriptionTerm>Gemaakt op</DescriptionTerm>
                        <DescriptionDetails>{data?.metadata?.createDate}</DescriptionDetails>
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
