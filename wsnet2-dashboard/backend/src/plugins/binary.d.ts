declare namespace binary {
  export type Unmarshaled = [unknown, object | null];
  export function UnmarshalRecursive(src: Uint8Array): Unmarshaled;
}

export = binary;
