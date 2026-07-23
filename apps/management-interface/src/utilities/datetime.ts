import {DateTime} from "luxon";

// Converts from the UTC time to the Europe/Amsterdam time.
export function formatDateTime(isoString: string): string {
    return DateTime.fromISO(isoString, {zone: 'utc'})
        .setZone('Europe/Amsterdam')
        .toFormat('dd-MM-yyyy HH:mm:ss');
}
