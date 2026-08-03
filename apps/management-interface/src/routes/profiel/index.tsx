import { createFileRoute } from '@tanstack/react-router'
import Card from '@/components/ui/card.tsx'
import { Badge } from '@/components/ui/badge.tsx'
import { DescriptionDetails, DescriptionList, DescriptionTerm } from '@/components/ui/description-list.tsx'
import { Heading } from '@/components/ui/heading.tsx'
import { Text } from '@/components/ui/text.tsx'
import { useProfile } from '@/auth/useProfile'
import { formatDateTime } from '@/utilities/datetime'

export const Route = createFileRoute('/profiel/')({
    component: RouteComponent,
})

const UNKNOWN = 'Niet beschikbaar'

function RouteComponent() {
    const { displayName, email, subject, issuer, expiresAt, roles } = useProfile()

    return (
        <div className="flex flex-col gap-6">
            <Heading>Profiel</Heading>
            <Card className="max-w-3xl">
                <DescriptionList>
                    <DescriptionTerm>Gebruikersnaam</DescriptionTerm>
                    <DescriptionDetails>{displayName}</DescriptionDetails>
                    <DescriptionTerm>E-mailadres</DescriptionTerm>
                    <DescriptionDetails>{email ?? UNKNOWN}</DescriptionDetails>
                    <DescriptionTerm>Rollen</DescriptionTerm>
                    <DescriptionDetails>
                        <div className="flex flex-wrap gap-2">
                            {roles.length > 0 ? (
                                // Role names are open, operator-defined data (ADR 0002) — render
                                // them as they come out of the token, without translating or
                                // mapping them onto a fixed set.
                                roles.map((role) => <Badge key={role} color="cyan">{role}</Badge>)
                            ) : (
                                <span>Geen rollen toegekend</span>
                            )}
                        </div>
                    </DescriptionDetails>
                    <DescriptionTerm>Principal</DescriptionTerm>
                    <DescriptionDetails className="font-mono">{subject ? `user::${subject}` : UNKNOWN}</DescriptionDetails>
                    <DescriptionTerm>Uitgegeven door</DescriptionTerm>
                    <DescriptionDetails>{issuer ?? UNKNOWN}</DescriptionDetails>
                    <DescriptionTerm>Sessie verloopt</DescriptionTerm>
                    <DescriptionDetails>
                        {expiresAt ? formatDateTime(new Date(expiresAt * 1000).toISOString()) : UNKNOWN}
                    </DescriptionDetails>
                </DescriptionList>
            </Card>
            <Text className="max-w-3xl">
                Deze gegevens komen uit uw inlogtoken en worden beheerd door uw
                identiteitsprovider; ze kunnen hier niet worden gewijzigd.
            </Text>
        </div>
    )
}
