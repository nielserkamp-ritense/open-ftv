import {createFileRoute} from '@tanstack/react-router'
import {FTVHeading} from "@/components/ui/heading.tsx";
import Card from "@/components/ui/card.tsx";
import {Textarea} from "@/components/ui/textarea.tsx";
import {DescriptionDetails, DescriptionList, DescriptionTerm} from "@/components/ui/description-list.tsx";
import {AuthlogEntry} from "@/services/authlog.ts";

export const Route = createFileRoute('/logboek/$id/')({
    component: RouteComponent,
})

function RouteComponent() {
    const navigate = Route.useNavigate()

    const entry = (history.state as { entry?: AuthlogEntry })?.entry

    // If no entry data is available, navigate back to overview
    if (!entry) {
        navigate({to: '/logboek'})
          .catch(e => console.error('Navigation error:', e));
        return null
    }

    // Pretty print JSON objects
    const responseJson = entry.response ? JSON.stringify(entry.response, null, 2) : ''
    const requestJson = entry.request ? JSON.stringify(entry.request, null, 2) : ''
    const informationJson = entry.information ? JSON.stringify(entry.information, null, 2) : ''
    const engineJson = entry.engine ? JSON.stringify(entry.engine, null, 2) : ''

    return <>
        <div className="flex items-center justify-between">
            <h1 className="text-rhc-color-cool-grey-900 font-normal text-[32px] leading-10">Overzicht logboek regel {entry.id}</h1>
        </div>
        <div className="flex flex-col 2xl:flex-row py-3 gap-6">
            <Card className="w-2/3 min-w-3xl flex-1 h-[836px]" disablePadding={true}>
                <div className="flex flex-col h-full px-4 py-5 sm:p-6 gap-4">
                    <div className="flex flex-col flex-1">
                        <FTVHeading level={2}>Response</FTVHeading>
                        <Textarea 
                            name="response" 
                            value={responseJson} 
                            readOnly={true} 
                            disabled={true} 
                            className="flex-1 bg-content-tertiary mt-3 font-mono text-sm"
                        />
                    </div>
                    <div className="flex flex-col flex-1">
                        <FTVHeading level={2}>Request</FTVHeading>
                        <Textarea 
                            name="request" 
                            value={requestJson} 
                            readOnly={true} 
                            disabled={true} 
                            className="flex-1 bg-content-tertiary mt-3 font-mono text-sm"
                        />
                    </div>
                </div>
            </Card>
            <div className="w-1/3 min-w-3xl 2xl:min-w-lg flex flex-col gap-6">
                <Card className="flex-1">
                    <DescriptionList>
                        <DescriptionTerm>ID</DescriptionTerm>
                        <DescriptionDetails>{entry.id}</DescriptionDetails>
                        
                        <DescriptionTerm>Tijdstempel</DescriptionTerm>
                        <DescriptionDetails>{entry.created}</DescriptionDetails>
                        
                        <DescriptionTerm>Request Type</DescriptionTerm>
                        <DescriptionDetails>{entry.requestType}</DescriptionDetails>
                        
                        <DescriptionTerm>Beleidsregels</DescriptionTerm>
                        <DescriptionDetails>{entry.policies ?? '-'}</DescriptionDetails>
                        
                        <DescriptionTerm>Trace ID</DescriptionTerm>
                        <DescriptionDetails>{entry.traceId ?? '-'}</DescriptionDetails>
                        
                        <DescriptionTerm>Span ID</DescriptionTerm>
                        <DescriptionDetails>{entry.spanId ?? '-'}</DescriptionDetails>
                        
                        {entry.information && Object.keys(entry.information).length > 0 && (
                            <>
                                <DescriptionTerm>Information</DescriptionTerm>
                                <DescriptionDetails>
                                    <pre className="text-xs overflow-auto max-h-40 font-mono">{informationJson}</pre>
                                </DescriptionDetails>
                            </>
                        )}
                        
                        {entry.engine && Object.keys(entry.engine).length > 0 && (
                            <>
                                <DescriptionTerm>Engine</DescriptionTerm>
                                <DescriptionDetails>
                                    <pre className="text-xs overflow-auto max-h-40 font-mono">{engineJson}</pre>
                                </DescriptionDetails>
                            </>
                        )}
                    </DescriptionList>
                </Card>
            </div>
        </div>
    </>
}
