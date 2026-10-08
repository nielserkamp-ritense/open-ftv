import { Zap, User } from "lucide-react";
import { Link } from "react-router-dom";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Button } from "@/components/ui/button";
import { useGebruikers } from "@/lib/api/gebruikers";

interface HeaderProps {
  currentUser?: string;
  onUserChange?: (user: string) => void;
}

// Rol mapping gebaseerd op afdeling
const getUserRole = (userName: string, gebruikers: Array<{ naam: string; afdeling: string }>): 'gemeentesecretaris' | 'ambtenaar' => {
  const gebruiker = gebruikers.find(g => g.naam === userName);
  return gebruiker?.afdeling.toLowerCase() === 'gemeentesecretaris' ? 'gemeentesecretaris' : 'ambtenaar';
};

export const Header = ({
  currentUser = "Morty",
  onUserChange,
}: HeaderProps) => {
  const { data: gebruikers = [], isLoading } = useGebruikers();
  const userRole = getUserRole(currentUser, gebruikers);

  return (
    <header className="bg-card border-b border-border shadow-sm">
      <div className="container mx-auto px-4 py-4">
        <div className="flex items-center justify-between">
          <Link to="/behandelaar" className="flex items-center gap-3 hover:opacity-80 transition-opacity">
            <div className="bg-primary p-2 rounded-lg">
              <Zap className="h-6 w-6 text-primary-foreground" />
            </div>
            <div>
              <h1 className="text-xl font-bold text-foreground">Gemeente Portal</h1>
              <p className="text-sm text-muted-foreground">Laadpaalvergunningen</p>
            </div>
          </Link>

          <div className="flex items-center gap-4">
            <div className="flex items-center gap-2">
              <User className="h-5 w-5 text-muted-foreground" />
              <Select value={currentUser} onValueChange={onUserChange} disabled={isLoading}>
                <SelectTrigger className="w-[260px]" aria-label="Gebruiker">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent className="bg-background z-50">
                  {gebruikers.map((gebruiker) => (
                    <SelectItem key={gebruiker.id} value={gebruiker.naam}>
                      {gebruiker.naam} ({gebruiker.afdeling})
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            
            {userRole === 'gemeentesecretaris' && (
              <div className="px-3 py-1 bg-primary/10 border border-primary/30 rounded-md">
                <p className="text-xs font-medium text-primary">Verhoogde autorisatie</p>
              </div>
            )}
          </div>
        </div>
      </div>
    </header>
  );
};
