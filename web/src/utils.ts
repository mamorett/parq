export function formatDate(value: any): string {
  if (value === null || value === undefined) return '';
  
  // If it's a number or bigint, assume it's a timestamp
  if (typeof value === 'number' || typeof value === 'bigint') {
    const num = Number(value);
    let date: Date;

    if (num > 1e16) { // Nanos
      date = new Date(num / 1e6);
    } else if (num > 1e13) { // Micros
      date = new Date(num / 1e3);
    } else if (num > 1e11) { // Millis
      date = new Date(num);
    } else { // Seconds
      date = new Date(num * 1000);
    }

    if (!isNaN(date.getTime()) && num >= 946684800) {
      return date.toLocaleString();
    }
  }

  // If it's a string, try to parse it
  if (typeof value === 'string') {
    const date = new Date(value);
    if (!isNaN(date.getTime()) && value.length > 5) {
        return date.toLocaleString();
    }
  }

  return String(value);
}
