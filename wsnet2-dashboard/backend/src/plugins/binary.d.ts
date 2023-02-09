declare namespace binary {
  export type Unmarshaled = [unknown, object | null];
  export function UnmarshalRecursive(src: Uint8Array): Unmarshaled;
  export function MarshalStr8(src: string): Uint8Array;
  export function MarshalFloat(src: number): Uint8Array;
  export function MarshalInt(src: number): Uint8Array;
  export function MarshalDict(src: object): Uint8Array;
  export function MarshalList(src: Array<Uint8Array>): Uint8Array;
  export function MarshalNull(): Uint8Array;
}

export = binary;
