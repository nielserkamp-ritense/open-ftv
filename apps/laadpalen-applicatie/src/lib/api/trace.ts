export const newTraceId = () => randomHex(16)

// Kong and the PDPs log their decisions under the trace ID of this traceparent, so the trace ID finds them in the
// Logboek.
export const traceparent = (traceId: string) => `00-${traceId}-${randomHex(8)}-01`

const randomHex = (bytes: number) =>
  Array.from(crypto.getRandomValues(new Uint8Array(bytes)), (b) => b.toString(16).padStart(2, '0')).join('')
