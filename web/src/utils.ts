export function formatDate(value: any): string {
  if (value === null || value === undefined) return 'N/A';
  
  // If it's a number, assume it's a timestamp (seconds or milliseconds)
  if (typeof value === 'number') {
    // If it's something like 1714800000 (seconds) or 1714800000000 (ms)
    const date = value > 1e11 ? new Date(value) : new Date(value * 1000);
    if (!isNaN(date.getTime())) {
      return date.toLocaleString();
    }
  }

  // If it's a string, try to parse it
  if (typeof value === 'string') {
    const date = new Date(value);
    if (!isNaN(date.getTime()) && value.length > 5) { // Basic check to avoid parsing short strings as dates
        return date.toLocaleString();
    }
  }

  return String(value);
}
