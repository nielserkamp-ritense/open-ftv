import {createFileRoute} from '@tanstack/react-router'
import {FTVHeading, Heading} from "@/components/ui/heading.tsx";
import Card from "@/components/ui/card.tsx";
import {useAttribute, useReplaceAttribute} from "@/services/attributes.ts";
import {ScaleLoader} from "react-spinners";
import {Textarea} from "@/components/ui/textarea.tsx";
import {DescriptionDetails, DescriptionList, DescriptionTerm} from "@/components/ui/description-list.tsx";
import { Badge } from "@/components/ui/badge.tsx";
import { badgeColorKeyFromString } from "@/utilities/color.ts";
import { formatDateTime } from "@/utilities/datetime.ts";
import { Button, SecondaryButton } from "@/components/ui/button.tsx";
import { useState } from "react";
import {Breadcrumb} from "@/components/ui/breadcrumb.tsx";

export const Route = createFileRoute('/attributen/$key/')({
    component: RouteComponent,
})

function RouteComponent() {
    const {key} = Route.useParams()
    const {status, data, error} = useAttribute(key)
    const replaceAttributeMutation = useReplaceAttribute()
    const [isEditing, setIsEditing] = useState(false)
    const [editedValue, setEditedValue] = useState<string>("")

    if (status == "pending") {
        return <>
            <div className="flex items-center justify-between">
                <Heading className="lg:text-3xl ">Attribuut</Heading>
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

    const handleEdit = () => {
        setEditedValue(data?.value as string ?? "")
        setIsEditing(true)
    }

    const handleSave = async () => {
        if (!data) return

        // Optimistically update UI
        setIsEditing(false)

        try {
            await replaceAttributeMutation.mutateAsync({
                key: key,
                attribute: {
                    ...data,
                    value: editedValue
                }
            })
        } catch (error) {
            // Revert on error
            setIsEditing(true)
            console.error('Failed to save attribute:', error)
        }
    }

    const displayValue = isEditing ? editedValue : (data?.value as string ?? "")

    return <>
        <Breadcrumb items={[
          { label: 'Home', href: '/' },
          { label: 'Context', href: '/attributen' },
          { label: data?.metadata?.title || 'Details' }
        ]}/>
        <div className="flex items-center justify-between">
            <h1 className="text-rhc-color-cool-grey-900 font-normal text-[32px] leading-10">Attribuut</h1>
            {isEditing ? (
                <Button color="primary"
                    onClick={handleSave}
                    disabled={replaceAttributeMutation.isPending}
                >
                    Opslaan
                </Button>
            ) : (
                <SecondaryButton onClick={handleEdit}>
                    Bewerken
                </SecondaryButton>
            )}
        </div>
        <div className="flex flex-col 2xl:flex-row py-3 gap-6">
            <Card className="w-2/3 min-w-3xl flex-1 h-[836px]" disablePadding={true}>
                <div className="flex flex-col h-full px-4 py-5 sm:p-6">
                    <FTVHeading level={2}>{data?.metadata.title}</FTVHeading>
                    <span>Waarde</span>
                    <Textarea
                        name="data"
                        value={displayValue}
                        onChange={(e) => setEditedValue(e.target.value)}
                        readOnly={!isEditing}
                        disabled={!isEditing}
                        className="flex-1 bg-content-tertiary mt-3"
                    />
                </div>
            </Card>
            <div className="w-1/3 min-w-3xl 2xl:min-w-lg flex flex-col gap-6">
                <Card className="flex-1">
                    <DescriptionList>
                      <DescriptionTerm>ID</DescriptionTerm>
                      <DescriptionDetails>{data?.key}</DescriptionDetails>
                        <DescriptionTerm>Omschrijving</DescriptionTerm>
                        <DescriptionDetails>{data?.metadata.title}</DescriptionDetails>
                        <DescriptionTerm>Datatype</DescriptionTerm>
                        <DescriptionDetails>{data?.type}</DescriptionDetails>
                        <DescriptionTerm>Gemaakt door</DescriptionTerm>
                        <DescriptionDetails>{data?.audit?.createdBy || "Onbekend"}</DescriptionDetails>
                        <DescriptionTerm>Gemaakt op</DescriptionTerm>
                        <DescriptionDetails>{data?.audit?.created ? formatDateTime(data.audit.created) : "Onbekend"}</DescriptionDetails>
                        <DescriptionTerm>Laatst bijgewerkt op</DescriptionTerm>
                        <DescriptionDetails>{data?.audit?.updated ? formatDateTime(data.audit.updated) : "Onbekend"}</DescriptionDetails>
                    </DescriptionList>
                </Card>
                <Card className="flex-1" header={
                    <Heading>Gebruik</Heading>
                }>
                  <DescriptionList>
                    <DescriptionTerm>Beslispunten</DescriptionTerm>
                    <DescriptionDetails>
                      <div className="flex flex-wrap gap-2">
                        {tags.length > 0 ? (
                          tags.map((tag) => (
                            <Badge key={tag} color={badgeColorKeyFromString(tag)}>{tag}</Badge>
                          ))
                        ) : (
                          <span>Geen beslispunten</span>
                        )}
                      </div>
                    </DescriptionDetails>

                  </DescriptionList>
                </Card>
            </div>
        </div>
    </>
}
