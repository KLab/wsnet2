import { h } from "vue";
import {
  NList,
  NListItem,
  NTag,
  NIcon,
  NCollapse,
  NCollapseItem,
} from "naive-ui";
import { Error as Errormark, Checkmark } from "@vicons/carbon";

export function renderBoolean(val: boolean) {
  if (val) {
    return h(
      NIcon,
      {
        color: "green",
      },
      { default: () => h(Checkmark) }
    );
  } else {
    return h(
      NIcon,
      {
        color: "red",
      },
      { default: () => h(Errormark) }
    );
  }
}

export function renderNumber(val: bigint | number) {
  return h(
    NTag,
    { type: "warning", round: true },
    { default: () => val.toString() }
  );
}

export function renderString(val: string) {
  return val;
}

export function renderEmpty() {
  return h(NTag, { disabled: true }, { default: () => "Null" });
}

export function renderObject(val: object | null) {
  if (val == null) return renderEmpty();
  if (val instanceof Array) {
    return h(
      NCollapse,
      {},
      {
        default: () =>
          h(
            NCollapseItem,
            { title: "" },
            {
              default: () =>
                h(
                  NList,
                  { bordered: true },
                  {
                    default: () =>
                      val.map((item) =>
                        h(NListItem, {}, { default: () => render(item) })
                      ),
                  }
                ),
            }
          ),
      }
    );
  } else if (val.constructor == Object) {
    return h(
      NCollapse,
      {},
      {
        default: () =>
          h(
            NCollapseItem,
            { title: "" },
            {
              default: () =>
                h(
                  NList,
                  { bordered: true },
                  {
                    default: () =>
                      Object.entries(val).map(([key, value]) =>
                        h(
                          NListItem,
                          {},
                          {
                            default: () => render(value),
                            prefix: () =>
                              h(NTag, { type: "info" }, { default: () => key }),
                          }
                        )
                      ),
                  }
                ),
            }
          ),
      }
    );
  } else {
    return renderEmpty();
  }
}

export function render(src: unknown) {
  switch (typeof src) {
    case "boolean":
      return renderBoolean(src);
    case "string":
      return renderString(src);
    case "bigint":
    case "number":
      return renderNumber(src);
    case "object":
      return renderObject(src);
    default:
      return renderEmpty();
  }
}
