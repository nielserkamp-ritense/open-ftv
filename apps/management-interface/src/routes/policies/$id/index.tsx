import {createFileRoute} from '@tanstack/react-router'
import {FTVHeading, Heading} from "@/components/ui/heading.tsx";
import Card from "@/components/ui/card.tsx";
import {usePolicy} from "@/services/policies.ts";
import {ScaleLoader} from "react-spinners";
import {Textarea} from "@/components/ui/textarea.tsx";
import {DescriptionDetails, DescriptionList, DescriptionTerm} from "@/components/ui/description-list.tsx";
import {Navbar, NavbarItem, NavbarSection} from "@/components/ui/navbar.tsx";
import {IconFileText, IconPlayerPlay, IconSettings, IconSourceCode} from "@tabler/icons-react";
import { Badge } from "@/components/ui/badge.tsx";
import { badgeColorKeyFromString } from "@/utilities/color.ts";

export const Route = createFileRoute('/policies/$id/')({
    component: RouteComponent,
})

function RouteComponent() {
    const {id} = Route.useParams()
    const {status, data, error} = usePolicy(id)

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

    const tags: string[] = data?.metadata?.tags ?? []

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
                            <NavbarItem href={"/policies/" + id + "/edit/" }>
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
                        <DescriptionDetails>
                            <div className="flex flex-wrap gap-2">
                                {tags.length > 0 ? (
                                    tags.map((tag) => (
                                        <Badge key={tag} color={badgeColorKeyFromString(tag)}>{tag}</Badge>
                                    ))
                                ) : (
                                    <span>Geen tags</span>
                                )}
                            </div>
                        </DescriptionDetails>
                        <DescriptionTerm>Gemaakt door</DescriptionTerm>
                        <DescriptionDetails>{data?.audit?.createdBy}</DescriptionDetails>
                        <DescriptionTerm>Laatst bijgewerkt</DescriptionTerm>
                        <DescriptionDetails>{data?.audit?.updated}</DescriptionDetails>
                        <DescriptionTerm>Gemaakt op</DescriptionTerm>
                        <DescriptionDetails>{data?.audit?.created}</DescriptionDetails>
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
