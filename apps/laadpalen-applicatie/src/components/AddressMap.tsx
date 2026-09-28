import { useEffect, useRef } from "react";
import L from "leaflet";
import "leaflet/dist/leaflet.css";

// Fix for default marker icons
delete (L.Icon.Default.prototype as { _getIconUrl?: unknown })._getIconUrl;
L.Icon.Default.mergeOptions({
  iconRetinaUrl: "https://unpkg.com/leaflet@1.9.4/dist/images/marker-icon-2x.png",
  iconUrl: "https://unpkg.com/leaflet@1.9.4/dist/images/marker-icon.png",
  shadowUrl: "https://unpkg.com/leaflet@1.9.4/dist/images/marker-shadow.png",
});

interface AddressMapProps {
  address: string;
  postcode: string;
}

// Realistische coördinaten langs straten in een fictieve wijk van Amsterdam
const postcodeCoordinates: Record<string, [number, number]> = {
  "1111AA": [52.36756, 4.90412],  // Korenbloemplein - hoek
  "1111AB": [52.36763, 4.90425],  // Korenbloemplein - langs straat
  "1111BA": [52.36772, 4.90438],  // Wilgenlaan - noordzijde
  "1111BB": [52.36765, 4.90441],  // Wilgenlaan - zuidzijde
  "1111CA": [52.36781, 4.90455],  // Esdoornstraat - oostzijde
  "1111CB": [52.36774, 4.90458],  // Esdoornstraat - westzijde
  "1111DA": [52.36790, 4.90472],  // Populierweg - noordzijde
  "1111DB": [52.36783, 4.90475],  // Populierweg - zuidzijde
  "1112AA": [52.36799, 4.90489],  // Lindenlaan - oostzijde
  "1112AB": [52.36792, 4.90492],  // Lindenlaan - westzijde
};

export const AddressMap = ({ address, postcode }: AddressMapProps) => {
  const mapRef = useRef<HTMLDivElement>(null);
  const mapInstanceRef = useRef<L.Map | null>(null);

  useEffect(() => {
    if (!mapRef.current || mapInstanceRef.current) return;

    // Initialize map met standaard startpositie
    const coordinates = postcodeCoordinates[postcode] || [52.36756, 4.90412];

    // Initialize map met betere zoom voor straatdetail
    const map = L.map(mapRef.current).setView(coordinates, 19);

    // Add OpenStreetMap tiles
    L.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", {
      attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
    }).addTo(map);

    // Add marker
    const marker = L.marker(coordinates).addTo(map);
    marker.bindPopup(`<strong>Locatie laadpaal</strong><br/>${address}`);

    mapInstanceRef.current = map;

    // Cleanup
    return () => {
      if (mapInstanceRef.current) {
        mapInstanceRef.current.remove();
        mapInstanceRef.current = null;
      }
    };
  }, [address, postcode]);

  return (
    <div 
      ref={mapRef} 
      className="h-64 w-full rounded-lg border border-border"
      style={{ zIndex: 0 }}
    />
  );
};

