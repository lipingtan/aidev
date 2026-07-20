declare module 'jsondiffpatch' {
  export function create(options?: any): any
  export const formatters: {
    html: {
      format(delta: any, left?: any): string
    }
  }
}
