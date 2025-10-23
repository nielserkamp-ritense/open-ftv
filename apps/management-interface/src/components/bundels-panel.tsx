import {useState, useEffect} from "react";
import {Button} from "@/components/ui/button.tsx";
import {IconLayoutSidebarRightCollapse, IconPlus} from "@tabler/icons-react";
import {Accordion, AccordionItem, AccordionTrigger, AccordionContent} from "@/components/ui/accordion.tsx";
import {Switch} from "@/components/ui/switch.tsx";
import Grid from "@/components/ui/grid.tsx";
import {useDeployments, useDeployment} from "@/services/bundles.ts";

export interface BundelsPanelProps {
    isOpen: boolean;
    onClose: () => void;
}

export default function BundelsPanel({isOpen, onClose}: BundelsPanelProps) {
    const {data, isLoading, isError} = useDeployments();
    // Track which accordion items are open to lazily load their data
    const [openItems, setOpenItems] = useState<string[]>([]);

    if (!isOpen) return null;

    return (
        <div className={"pt-6 pb-8 px-10 gap-8 flex-1"}>
            <div className="flex justify-between items-center">
                <div>
                    <span className={"text-rhc-lintblauw-500 text-[20px] font-normal"}>Bundels</span>
                </div>
                <div>
                    <button
                        onClick={onClose}
                        className="flex items-center gap-1 cursor-pointer hover:opacity-70 transition-opacity"
                    >
            <span className={"text-carrotnl-button-subtle-color text-[14px] font-semibold"}>
              Verberg bundels
            </span>
                        <IconLayoutSidebarRightCollapse
                            size={16}
                            className={`transition-transform ${isOpen ? "rotate-180" : ""}`}
                        />
                    </button>
                </div>
            </div>

            <div className={"pt-6 text-right"}>
                <Button disabled color={"primary"} href={"/policies/add"}>
                    <div className={"flex align-middle justify-center my-auto"}>
                        <IconPlus className={"text-content-inverse-secondary"} size={20}></IconPlus>
                    </div>
                    <span>Aanmaken bundel</span>
                </Button>
            </div>

            {/* Filter switch section */}
            <div className="pt-8 h-[48px]">
                <div className="flex items-center gap-3">
                    <Switch disabled aria-label="Filter op bundels waar geselecteerde regel in voorkomt"/>
                    <span className="text-[16px] text-content-primary">Filter op bundels waar geselecteerde regel in voorkomt</span>
                </div>
            </div>

            {/* Accordions replacing previous content area */}
            <div className="pt-8 w-full mt-4">
                {isLoading && (
                    <div className="text-content-secondary text-sm">Laden van bundels...</div>
                )}
                {isError && (
                    <div className="text-danger-600 text-sm">Er is een fout opgetreden bij het laden van bundels.</div>
                )}
                {!isLoading && !isError && (
                    <Accordion
                        type="multiple"
                        className="w-full border-t border-t-['#DEE2E6']"
                        value={openItems}
                        onValueChange={(val) => setOpenItems(Array.isArray(val) ? val : [val])}
                    >
                        {data && data.length > 0 ? (
                            data.map((b) => (
                                <AccordionItem key={b.version} value={`bundel-${b.version}`}>
                                    <AccordionTrigger className={"text-rhc-lintblauw-500 text-[18px]"}>
                                        {b.title || `Bundel ${b.version}`}
                                    </AccordionTrigger>
                                    <AccordionContent>
                                        <DeploymentDetails version={b.version}
                                                           isOpen={openItems.includes(`bundel-${b.version}`)}/>
                                    </AccordionContent>
                                </AccordionItem>
                            ))
                        ) : (
                            <div className="text-content-secondary text-sm py-4">Geen bundels gevonden.</div>
                        )}
                    </Accordion>
                )}
            </div>
        </div>
    );
}

function DeploymentDetails({version, isOpen}: { version: number; isOpen: boolean }) {
    // Delay enabling the query for 400ms after the accordion opens
    const [enabled, setEnabled] = useState(false);
    useEffect(() => {
        let timer: number | undefined;
        if (isOpen) {
            timer = window.setTimeout(() => setEnabled(true), 400);
        } else {
            setEnabled(false);
        }
        return () => {
            if (timer) window.clearTimeout(timer);
        };
    }, [isOpen]);

    const {data, isLoading, isError} = useDeployment(enabled ? version : undefined);

    if (!isOpen) return null;

    if (isLoading) {
        return <div className="text-content-secondary text-sm py-2">Laden…</div>;
    }
    if (isError) {
        return <div className="text-danger-600 text-sm py-2">Kon deployment niet laden.</div>;
    }

    if (!data) {
        return
    }

    const columnDefs = [
        {headerName: "Policy ID", field: "policyId", flex: 1},
        {headerName: "Titel", field: "titel", flex: 1},
        {headerName: "Regeling", field: "regeling", flex: 1},
        {
            headerName: "Actief",
            field: "actief",
            // eslint-disable-next-line @typescript-eslint/no-explicit-any
            valueFormatter: (p: any) => (p.value ? "Ja" : "Nee"),
            flex: 0.6,
        },
        {headerName: "Laatst bewerkt", field: "laatstBewerkt", flex: 1},
        {headerName: "Beslispunt", field: "beslispunt", flex: 1},
    ];

    return (
        <div className="py-2">
            <div className="h-[280px] w-full">
                <Grid
                    columnDefs={columnDefs}
                    data={data}
                    domLayout="autoHeight"
                    suppressCellFocus={true}
                />
            </div>
        </div>
    );
}