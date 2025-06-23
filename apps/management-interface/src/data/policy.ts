import { components } from "@/oas/policies"

export function GetPolicies() : components["schemas"]["Policy"][] {
    return [
        {
            id: "1234",
            language: "opa",
            url: "https://some.site/policies/1234",
            rvvaId: "e3"
        },
        {
            id: "1236",
            language: "opa",
            url: "https://some.site/policies/1236",
            rvvaId: "e3"
        },
        {
            id: "1278",
            language: "opa",
            url: "https://some.site/policies/1278",
            rvvaId: "e3"
        },
        {
            id: "1301",
            language: "opa",
            url: "https://some.site/policies/1301",
            rvvaId: "e3"
        },
    ]
}
